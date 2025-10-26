package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	ui "github.com/Mojjedrengen/ChitChat/client/UI"
	messageclient "github.com/Mojjedrengen/ChitChat/client/messageClient"
	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	serverAddr = flag.String("addr", "localhost:50051", "The server adress in the format of host:port")
	dataFolder = flag.String("folder", "data/", "the folder for where the log is saved")
)

func main() {
	user := ui.BasicLogin()
	username := user.Uuid
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", *dataFolder, username), os.ModePerm); err != nil {
		log.Fatalf("failed to creat dirr: %v", err)
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/%s/client.log", *dataFolder, username), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failted to open file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	fmt.Println("hello world")

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	conn, err := grpc.NewClient(*serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := chitchat.NewChatClient(conn)
	//user := &chitchat.User{
	//	Uuid: fmt.Sprintf("user-%d", rand.Intn(1000)),
	//}
	messageClient := messageclient.NewClient(user, client)
	reciveBuffer := make(chan *chitchat.Msg, 10)
	sendBuffer := make(chan string, 5)
	go messageClient.Connect(reciveBuffer)
	go messageClient.SendMessage(sendBuffer)
	ui.SetUpUI(reciveBuffer, sendBuffer, messageClient, user)

	for {

	}
}
