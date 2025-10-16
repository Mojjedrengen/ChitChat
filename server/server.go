package main

import (
	"context"
	"fmt"

	pb "github.com/Mojjedrengen/ChitChat/grpc"
)

func main() {
	fmt.Println("hello world")
}

type ChatServer struct {
	pb.UnimplementedChatServer

	MessageHistory []*pb.Msg
}

func (s *ChatServer) Connect(ctx context.Context, msg *pb.SimpleMessage) (*pb.ConnectRespond, error) {

}
