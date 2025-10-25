package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	pb "github.com/Mojjedrengen/ChitChat/grpc"
	"github.com/Mojjedrengen/ChitChat/util"
	"google.golang.org/grpc"
)

var (
	port       = flag.Int("port", 50051, "The Server port")
	file       = flag.String("file", "data.json", "Path to the file to store chat messages")
	dataFolder = flag.String("folder", "data", "the folder for where the log is saved")
)

type ChatServer struct {
	pb.UnimplementedChatServer

	MessageHistory             []*pb.Msg
	ConnectedClientsOut        map[*pb.User](chan *pb.Msg)
	ConnectedClients           map[*pb.User](chan *pb.Msg)
	ConnectedClientsDisconnect map[*pb.User](chan bool)
	DisconnectClientRespons    map[*pb.User](chan bool)
	mu                         sync.Mutex // for LastMessageIndex
	LastMessageIndex           map[*pb.User]int
	lamportClock               *util.LamportClock
}

var AdminUser = &pb.User{
	Uuid: "System",
}

func (s *ChatServer) Connect(msg *pb.SimpleMessage, stream pb.Chat_ConnectServer) error {
	user := msg.User
	message := msg.Message

	if message == "Connect" && user != nil {
		connectRespond := pb.ConnectRespond{
			StatusCode: &pb.ChatRespond{
				StatusCode: 200,
				Context:    "Connected",
			},
		}
		stream.Send(&connectRespond)
		s.mu.Lock()
		s.LastMessageIndex[user] = 0
		s.ConnectedClients[user] = make(chan *pb.Msg, 10)
		s.ConnectedClientsOut[user] = make(chan *pb.Msg, 10)
		s.ConnectedClientsDisconnect[user] = make(chan bool, 2)
		s.mu.Unlock()

		logicalTime := s.lamportClock.Tick()
		currTime := time.Now().Unix()

		connectedMsg := &pb.Msg{
			User:        AdminUser,
			UnixTime:    currTime,
			LogicalTime: logicalTime,
			Message:     fmt.Sprintf("participant %s joined Chit Chat at logical time %d", user.Uuid, logicalTime),
			Error:       fmt.Sprintf("participant %s have succesfully joined chat", user.Uuid),
		}
		log.Printf("CLIENT - %s Connect joined at logical time %v", user.Uuid, logicalTime)

		s.mu.Lock() //Same as other adminuser lock
		s.ConnectedClients[AdminUser] <- connectedMsg
		for _, msg := range s.MessageHistory {
			stream.Send(&pb.ConnectRespond{
				StatusCode: &pb.ChatRespond{
					StatusCode: 100,
					Context:    "Broadcasting old messages",
				},
				Message: msg,
			})
		}
		s.mu.Unlock()
		for {
			select {
			case isDisconnected := <-s.ConnectedClientsDisconnect[user]:
				if isDisconnected {
					stream.Send(&pb.ConnectRespond{
						StatusCode: &pb.ChatRespond{
							StatusCode: 100,
							Context:    "Connection is now Disconnected",
						},
						Message: &pb.Msg{
							User:        AdminUser,
							UnixTime:    time.Now().Unix(),
							LogicalTime: s.lamportClock.GetTime(),
							Message:     "Client is disconnected",
							Error:       "Disconnected",
						},
					})
					s.DisconnectClientRespons[user] <- true
					return nil
				}
			default:

				s.mu.Lock() //Bufferhandler might iterate over connectedclientsout so keep lock here
				outCh := s.ConnectedClientsOut[user]
				s.mu.Unlock()

				select {
				case msg := <-outCh:
					respond := pb.ConnectRespond{
						StatusCode: &pb.ChatRespond{
							StatusCode: 100,
							Context:    "Broadcast message",
						},
						Message: msg,
					}
					if err := stream.Send(&respond); err != nil {
						return err
					}
				default:
					time.Sleep(10 * time.Millisecond)
				}
			}
		}
	} else {
		stream.Send(&pb.ConnectRespond{
			StatusCode: &pb.ChatRespond{
				StatusCode: 401,
				Context:    "Use Valid connection",
			},
		})
		return errors.New("Invalid Authentication")
	}
}

func (s *ChatServer) OnGoingChat(stream pb.Chat_OnGoingChatServer) error {
	var userpb *pb.User

	first, err := stream.Recv()
	if err != nil {
		return err
	}
	userpb = first.User
	//use exact pointer stored in connect()
	var storedUser *pb.User
	s.mu.Lock()
	for u := range s.ConnectedClients {
		if u.Uuid == userpb.Uuid {
			storedUser = u
			break
		}
	}
	s.mu.Unlock()

	if storedUser == nil {
		return errors.New("User not found")
	}
	//Process the first message that was used for identification
	if first.Message != "" {
		timestamp := time.Now().Unix()
		logicalTime := s.lamportClock.Tick()

		if len(first.Message) >= 128 {
			stream.Send(&pb.ChatRespond{StatusCode: 401, Context: "ERROR: ILLEGAL LENGTH"})
		} else {
			sendMsg := &pb.Msg{
				User:        storedUser,
				Message:     first.Message,
				UnixTime:    timestamp,
				LogicalTime: logicalTime,
			}
			s.ConnectedClients[storedUser] <- sendMsg
			stream.Send(&pb.ChatRespond{StatusCode: 200, Context: "Message Send"})
		}
	}
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		message := in.Message
		timestamp := time.Now().Unix()
		s.mu.Lock() //lock here given that other goroutines might be writing to it for protection
		clientCh, ok := s.ConnectedClients[storedUser]
		s.mu.Unlock()
		if !ok {
			stream.Send(&pb.ChatRespond{StatusCode: 401, Context: "CONNECTION NOT ESTABLISHED"})
			return errors.New("unauthorized")
		}

		if len(message) >= 128 {
			stream.Send(&pb.ChatRespond{StatusCode: 401, Context: "ERROR: ILLEGAL LENGTH"})
			continue
		}

		logicalTime := s.lamportClock.Tick()

		sendMsg := &pb.Msg{
			User:        storedUser,
			Message:     message,
			UnixTime:    timestamp,
			LogicalTime: logicalTime,
		}
		clientCh <- sendMsg

		stream.Send(&pb.ChatRespond{StatusCode: 200, Context: "Message Send"})
	}

}

func (s *ChatServer) Disconnect(ctx context.Context, msg *pb.SimpleMessage) (*pb.ChatRespond, error) {
	message := msg.Message
	var storedUser *pb.User
	s.mu.Lock()
	for u := range s.ConnectedClients {
		if u.Uuid == msg.User.Uuid {
			storedUser = u
			break
		}
	}
	s.mu.Unlock()

	if storedUser == nil {
		return &pb.ChatRespond{
			StatusCode: 401,
			Context:    "ERROR: USER NOT FOUND",
		}, errors.New("user not found")
	}

	user := storedUser

	if message == "Disconnect" && user != nil {
		s.mu.Lock()
		s.DisconnectClientRespons[user] = make(chan bool, 2)
		for i := 0; i < 2; i++ {

			s.ConnectedClientsDisconnect[user] <- true

		}
		s.mu.Unlock()

		<-s.DisconnectClientRespons[user]

		s.mu.Lock()
		close(s.ConnectedClients[user])
		close(s.ConnectedClientsDisconnect[user])
		close(s.DisconnectClientRespons[user])
		delete(s.ConnectedClients, user)
		delete(s.ConnectedClientsDisconnect, user)
		delete(s.DisconnectClientRespons, user)
		delete(s.LastMessageIndex, user)
		s.mu.Unlock()

		// Increment Lamport clock for disconnect event
		logicalTime := s.lamportClock.Tick()
		time := time.Now().Unix()

		disconnectMsg := &pb.Msg{
			User:        AdminUser,
			UnixTime:    time,
			LogicalTime: logicalTime,
			Message:     fmt.Sprintf("participant %s left Chit Chat at logical time %d", user.Uuid, logicalTime),
			Error:       fmt.Sprintf("participant %s have succesfully left chat", user.Uuid),
		}
		log.Printf("CLIENT - %s Disconnect left at logical time %v", user.Uuid, logicalTime)
		s.mu.Lock() //keep lock here in case adminuser gets removed from map whilst sending
		s.ConnectedClients[AdminUser] <- disconnectMsg
		s.mu.Unlock()

		disconnectRespond := &pb.ChatRespond{
			StatusCode: 200,
			Context:    fmt.Sprintf("participant %s have succesfully left chat", user.Uuid),
		}

		return disconnectRespond, nil

	} else {

		disconnectRespond := &pb.ChatRespond{
			StatusCode: 400,
			Context:    "ERROR: INVALID DISCONNECT REQUEST",
		}
		return disconnectRespond, errors.New("err: invalid disconnect request")
	}
}

func bufferhandler(s *ChatServer) {
	var messageBuffer []*pb.Msg

	for {
		messageBuffer = []*pb.Msg{}
		s.mu.Lock() //Lock here since connect and disconnect are modifying map
		for _, channel := range s.ConnectedClients {
			select {
			case msg := <-channel:
				messageBuffer = append(messageBuffer, msg)
			default:
				continue
			}
		}
		s.mu.Unlock()
		messageBuffer = util.SortMsgListBasedOnTime(messageBuffer)

		s.mu.Lock()
		s.MessageHistory = append(s.MessageHistory, messageBuffer...)

		for _, msg := range messageBuffer {
			fmt.Printf("<%v @ %v (L:%v)> %v\n", msg.User.Uuid, msg.UnixTime, msg.LogicalTime, msg.Message)
			log.Printf("SERVER: Delivery from %v @ %v (Lamport: %v): %v", msg.User.Uuid, msg.UnixTime, msg.LogicalTime, msg.Message)
			for user, ch := range s.ConnectedClientsOut {
				if user == AdminUser {
					continue
				}
				select {
				case ch <- msg:
				default:
				}
			}
		}
		s.mu.Unlock()
	}
}

func newServer() *ChatServer {
	s := &ChatServer{
		MessageHistory:             make([]*pb.Msg, 0, 10),
		ConnectedClientsOut:        make(map[*pb.User]chan *pb.Msg),
		ConnectedClients:           make(map[*pb.User]chan *pb.Msg),
		ConnectedClientsDisconnect: make(map[*pb.User]chan bool),
		DisconnectClientRespons:    make(map[*pb.User]chan bool),
		LastMessageIndex:           make(map[*pb.User]int),
		lamportClock:               util.NewLamportClock(),
	}
	s.ConnectedClients[AdminUser] = make(chan *pb.Msg, 10)
	{ //Openens saved data
		file, err := os.OpenFile(fmt.Sprintf("%s/server/%s", *dataFolder, *file), os.O_CREATE|os.O_RDONLY, 0666)
		if err != nil {
			log.Println(err)
		}
		defer file.Close()

		var mh []*pb.Msg
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&mh); err != nil {
			log.Println(err)
		}
		s.mu.Lock()
		s.MessageHistory = mh
		fmt.Println("OLD MESSAGES:")
		for _, msg := range s.MessageHistory {
			fmt.Printf("<%v @ %v (L:%v)> %v\n", msg.User.Uuid, msg.UnixTime, msg.LogicalTime, msg.Message)
			if msg.LogicalTime > s.lamportClock.GetTime() {
				s.lamportClock.Update(msg.LogicalTime)
			}
		}
		s.mu.Unlock()
		fmt.Println("OLD MESSAGES DONE")
	}
	go bufferhandler(s)

	go func() { //function for saving data
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		file, err := os.OpenFile(fmt.Sprintf("%s/server/%s", *dataFolder, *file), os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Panic(err)
		}
		defer file.Close()

		s.mu.Lock()
		encoder := json.NewEncoder(file)
		if err := encoder.Encode(s.MessageHistory); err != nil {
			log.Panic(err)
		}
		log.Printf("SERVER: shutdown @ %v (Lamport: %v)", time.Now().Format(time.DateTime), s.lamportClock.GetTime())
		fmt.Print("\n")
		os.Exit(0)
	}()

	return s
}

func main() {
	if err := os.MkdirAll(fmt.Sprintf("%s/server", *dataFolder), os.ModePerm); err != nil {
		log.Fatalf("failed to creat dirr: %v", err)
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/server/server.log", *dataFolder), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failted to open file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	fmt.Printf("port: %v\n", *port)
	log.Printf("SERVER: Startup listening on port %v @ %s", *port, time.Now().Format(time.DateTime))

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}

