package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/Mojjedrengen/ChitChat/grpc"
	"github.com/Mojjedrengen/ChitChat/util"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The Server port")
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

		currTime := time.Now().Unix()
		connectedMsg := &pb.Msg{
			User:     AdminUser,
			UnixTime: currTime,
			Message:  fmt.Sprintf("participant %s joined Chit Chat at logical time %d", user.Uuid, currTime),
			Error:    fmt.Sprintf("participant %s have succesfully joined chat", user.Uuid),
		}
		s.mu.Lock()
		s.ConnectedClients[AdminUser] <- connectedMsg
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
							User:     AdminUser,
							UnixTime: time.Now().Unix(),
							Message:  "Client is disconnected",
							Error:    "Disconnected",
						},
					})
					s.DisconnectClientRespons[user] <- true
					return errors.New("Disconnected by user")
				}
			default:

				s.mu.Lock()
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

		s.mu.Lock()
		clientCh, ok := s.ConnectedClients[userpb]
		s.mu.Unlock()

		if !ok {
			stream.Send(&pb.ChatRespond{StatusCode: 401, Context: "CONNECTION NOT ESTABLISHED"})
			return errors.New("unauthorized")
		}

		if len(message) >= 128 {
			stream.Send(&pb.ChatRespond{StatusCode: 401, Context: "ERROR: ILLEGAL LENGTH"})
			continue
		}

		sendMsg := &pb.Msg{User: userpb, Message: message, UnixTime: timestamp}
		s.mu.Lock()
		clientCh <- sendMsg
		s.mu.Unlock()

		stream.Send(&pb.ChatRespond{StatusCode: 200, Context: "Message Send"})
	}

}

func (s *ChatServer) Disconnect(ctx context.Context, msg *pb.SimpleMessage) (*pb.ChatRespond, error) {
	user := msg.User
	message := msg.Message

	if message == "Disconnect" && user != nil {
		s.mu.Lock()
		s.DisconnectClientRespons[user] = make(chan bool, 2)
		s.mu.Unlock()
		for i := 0; i < 2; i++ {
			s.mu.Lock()
			s.ConnectedClientsDisconnect[user] <- true
			s.mu.Unlock()
		}
		responds := 0
		fullyDisconnected := false
		for fullyDisconnected {
			s.mu.Lock()
			if <-s.DisconnectClientRespons[user] {
				responds++
				if responds == 2 {
					fullyDisconnected = true
				}
			}
			s.mu.Unlock()
		}

		s.mu.Lock()
		close(s.ConnectedClients[user])
		close(s.ConnectedClientsDisconnect[user])
		close(s.DisconnectClientRespons[user])
		delete(s.ConnectedClients, user)
		delete(s.ConnectedClientsDisconnect, user)
		delete(s.DisconnectClientRespons, user)
		delete(s.LastMessageIndex, user)
		s.mu.Unlock()

		time := time.Now().Unix()
		disconnectMsg := &pb.Msg{
			User:     AdminUser,
			UnixTime: time,
			Message:  fmt.Sprintf("participant %s left Chit Chat at logical time %d", user.Uuid, time),
			Error:    fmt.Sprintf("participant %s have succesfully left chat", user.Uuid),
		}
		s.mu.Lock()
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
		s.mu.Lock()
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
	}
	s.ConnectedClients[AdminUser] = make(chan *pb.Msg, 10)
	go bufferhandler(s)
	return s
}

func main() {
	fmt.Printf("port: %v\n", *port)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterChatServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
