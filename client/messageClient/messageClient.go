package messageclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
)

type MessageClient struct {
	user           *chitchat.User
	messageHistroy []*chitchat.Msg
	messageLog     []*chitchat.ChatRespond
	mu             sync.Mutex
	client         chitchat.ChatClient
	interrupt      chan (os.Signal)
}

func NewClient(user *chitchat.User, client chitchat.ChatClient) *MessageClient {
	returnClient := MessageClient{
		messageHistroy: make([]*chitchat.Msg, 0),
		messageLog:     make([]*chitchat.ChatRespond, 0),
		user:           user,
		client:         client,
		interrupt:      make(chan os.Signal),
	}
	signal.Notify(returnClient.interrupt, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-returnClient.interrupt
		fmt.Printf("\n")
		returnClient.Disconenct()
	}()

	return &returnClient
}

func (messageClient *MessageClient) Connect(messageBuf chan<- *chitchat.Msg) {
	connectMessage := &chitchat.SimpleMessage{
		User:    messageClient.user,
		Message: "Connect",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := messageClient.client.Connect(ctx, connectMessage)
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
		messageClient.messageHistroy = append(messageClient.messageHistroy, message.Message)
		messageClient.messageLog = append(messageClient.messageLog, message.StatusCode)
		messageClient.mu.Unlock()
		if message.Message.User.Uuid == messageClient.user.Uuid {
			continue
		}
		messageBuf <- message.Message
	}
}

func (messageClient *MessageClient) SendMessage(messageChan <-chan string) {
	stream, err := messageClient.client.OnGoingChat(context.Background())
	if err != nil {
		log.Fatalf("Fail to establish send connection: %v", err)
	}
	for {
		msg := chitchat.SimpleMessage{
			User:    messageClient.user,
			Message: <-messageChan,
		}
		if err := stream.Send(&msg); err != nil {
			log.Fatalf("Failed to send message: %v", err)
		}
		resiveMessage, err := stream.Recv()
		if err != nil {
			log.Fatalf("Fail to get respond after message: %v", err)
		}
		messageClient.mu.Lock()
		messageClient.messageLog = append(messageClient.messageLog, resiveMessage)
		messageClient.mu.Unlock()
	}
}

func (messageClient *MessageClient) Disconenct() {
	disconnectMessage := &chitchat.SimpleMessage{
		User:    messageClient.user,
		Message: "Disconnect",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	respond, err := messageClient.client.Disconnect(ctx, disconnectMessage)
	if err != nil {
		log.Fatalf("client.disconnect failed: %v", err)
	}

	messageClient.mu.Lock()
	messageClient.messageLog = append(messageClient.messageLog, respond)
	messageClient.mu.Unlock()
	if respond.StatusCode != 200 {
		log.Fatalf("client.disconenct failed: %v", respond.Context)
	} else {
		os.Exit(0)
	}
}
