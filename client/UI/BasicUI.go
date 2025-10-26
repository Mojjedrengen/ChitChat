package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	termcommands "github.com/Mojjedrengen/ChitChat/client/Termcommands"
	messageclient "github.com/Mojjedrengen/ChitChat/client/messageClient"
	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
)

type BasicUI struct {
	mc           *messageclient.MessageClient
	reciveBuffer chan *chitchat.Msg
	sendBuffer   chan string
	helpMessage  string
	username     *chitchat.User
}

func SetUpUI(reciveBuffer chan *chitchat.Msg, sendBuffer chan string, messageClient *messageclient.MessageClient, user *chitchat.User) {
	UI := BasicUI{
		mc:           messageClient,
		reciveBuffer: reciveBuffer,
		sendBuffer:   sendBuffer,
		helpMessage:  "Write say {message} to write a message. Write Disconnect to disconnect from the server",
		username:     user,
	}
	go UI.printer()
	go UI.writer()
}

func (UI *BasicUI) printer() {
	for {
		fmt.Printf("%s<%s> ", termcommands.ResetColour, UI.username.Name)
		msg := <-UI.reciveBuffer

		if msg.User.Uuid == UI.username.Uuid {
			fmt.Printf("%s%s", termcommands.LineUp, termcommands.ClearLine)
		}
		fmt.Printf("%s", termcommands.ClearLine)
		// Display logical timestamp alongside physical timestamp
		switch msg.User.Uuid {
		case UI.username.Uuid:
			printMsg(msg, termcommands.OwnTextColour)
		case "System":
			printMsg(msg, termcommands.ServerTextColour)
		default:
			printMsg(msg, termcommands.OtherTextColour)
		}
	}
}

func printMsg(msg *chitchat.Msg, colour termcommands.TermComands) {
	outTime := time.Unix(msg.UnixTime, 0).Format(time.DateTime)
	fmt.Printf("%s<%s @ %s (L:%d)> %s%s\n", colour, msg.User.Name, outTime, msg.LogicalTime, msg.Message, termcommands.ResetColour)
}

func (UI *BasicUI) writer() {
	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error in reading user input: %v", err)
		}
		input = strings.TrimRight(input, "\n")
		input = strings.TrimSpace(input)

		UI.sendBuffer <- input
		fmt.Printf("<%s> ", UI.username.Name)
	}
}
