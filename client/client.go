package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"sync"

	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server adress in the format of host:port")
)

type MessageClient struct {
	user           chitchat.User
	messageHistroy []chitchat.Msg
	messageBuffer  []chitchat.Msg
	messageLog     []chitchat.ChatRespond
	mu             sync.Mutex
}

func newClient(user chitchat.User) *MessageClient {
	return &MessageClient{
		messageHistroy: make([]chitchat.Msg, 0),
		messageLog:     make([]chitchat.ChatRespond, 0),
		messageBuffer:  nil,
		user:           user,
	}
}

func connect(client chitchat.ChatClient, messageClient MessageClient) {
	connectMessage := &chitchat.SimpleMessage{
		User:    &messageClient.user,
		Message: "Connect",
	}
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.Connect(ctx, connectMessage)
	if err != nil {
		log.Fatalf("client.connect failed: %v", err)
	}
	respond, err := stream.Recv()
	if respond.StatusCode.StatusCode != 200 {
		log.Fatalf("client.connect failed: %v", respond.StatusCode.Context)
		log.Fatalf("client.connect failed: %v", err)
		cancel()
		return
	}

	for {
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("client.connect failed: %v", err)
		}
		messageClient.mu.Lock()
		messageClient.messageHistroy = append(messageClient.messageHistroy, *message.Message)
		if messageClient.messageBuffer == nil {
			messageClient.messageBuffer = make([]chitchat.Msg, 0, 1)
		}
		messageClient.messageBuffer = append(messageClient.messageBuffer, *message.Message)
		messageClient.mu.Unlock()
	}
}

func main() {
	fmt.Println("hello world")

	var opts []grpc.DialOption

	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := chitchat.NewChatClient(conn)
	user := chitchat.User{
		Uuid: "test",
	}
	messageClient := newClient(user)
}
