package main

import (
	"flag"
	"fmt"
	"log"

	messageclient "github.com/Mojjedrengen/ChitChat/client/messageClient"
	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
	"google.golang.org/grpc"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server adress in the format of host:port")
)

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
	messageClient := messageclient.NewClient(user)
}
