package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	pb "github.com/Mojjedrengen/ChitChat/grpc"
)

func main() {
	fmt.Println("hello world")
}

type ChatServer struct {
	pb.UnimplementedChatServer

	MessageHistory             []*pb.Msg
	ConnectedClients           map[*pb.User](chan *pb.Msg)
	ConnectedClientsDisconnect map[*pb.User](chan bool)
	mu                         sync.Mutex // for LastMessageIndex
	LastMessageIndex           map[*pb.User]int
}

func (s *ChatServer) Connect(ctx context.Context, msg *pb.SimpleMessage) (*pb.ConnectRespond, error) {
	return nil, errors.New("Not implemented")
}

func (s *ChatServer) OnGoingChat(stream pb.Chat_OnGoingChatServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		user := in.User
		message := in.Message
		timestamp := time.Now().Unix()

		sendMessage := &pb.Msg{
			User:     user,
			Message:  "",
			UnixTime: timestamp,
			Error:    "",
		}
		var respond pb.ChatRespond

		if len(message) >= 128 {
			sendMessage.Error = "ILLEGAL LENGTH"
			respond = pb.ChatRespond{
				StatusCode: 400,
				Context:    "ERROR: ILLEGAL LENGTH OF MESSAGE, MESSAGE CANNOT EXCEED 128",
			}
		} else {
			sendMessage.Message = message
			s.ConnectedClients[user] <- sendMessage
			respond = pb.ChatRespond{
				StatusCode: 200,
				Context:    "Message Send",
			}
		}
		if err := stream.Send(&respond); err != nil {
			return err
		}
		/*
			var lastMessageIndex int
			for lastMessageIndex = s.LastMessageIndex[user]; lastMessageIndex <= len(s.MessageHistory)-1; lastMessageIndex++ {
				if err := stream.Send(s.MessageHistory[lastMessageIndex]); err != nil {
					return err
				}
			}
			s.mu.Lock()
			s.LastMessageIndex[user] = lastMessageIndex
			s.mu.Unlock()
		*/
	}
}
