package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	messageclient "github.com/Mojjedrengen/ChitChat/client/messageClient"
	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
)

type BasicUI struct {
	mc           messageclient.MessageClient
	reciveBuffer chan *chitchat.Msg
	sendBuffer   chan string
	helpMessage  string
}

func SetUpUI(reciveBuffer chan *chitchat.Msg, sendBuffer chan string, messageClient messageclient.MessageClient) {
	UI := BasicUI{
		mc:           messageClient,
		reciveBuffer: reciveBuffer,
		sendBuffer:   sendBuffer,
		helpMessage:  "Write say {message} to write a message. Write Disconnect to disconnect from the server",
	}
	go UI.printer()
	go UI.writer()
}

func (UI BasicUI) printer() {
	for {
		fmt.Print("> ")
		msg := <-UI.reciveBuffer
		out := strings.Builder{}
		out.WriteString("<")

		outTime := time.Unix(msg.UnixTime, 0).Format(time.DateTime)
		fmt.Printf("<%s @ %s> %s\n", msg.User, outTime, msg.Message)
	}
}

func (UI BasicUI) writer() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error in reading user input: %v", err)
		}
		input = strings.TrimRight(input, "\n")
		input = strings.TrimSpace(input)

		commands := strings.SplitN(input, " ", 2)
		commands[2] = strings.TrimSpace(commands[2])

		if commands[1] == "Disconnect" {
			UI.mc.Disconenct()
		} else if commands[1] == "Say" && len(commands[2]) != 0 {
			UI.sendBuffer <- commands[2]
			fmt.Print("\n> ")
		} else {
			fmt.Println(UI.helpMessage)
			fmt.Print("> ")
		}
	}
}
