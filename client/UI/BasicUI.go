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

type TermComands string

const (
	ClearLine        TermComands = "\033[2K\r"
	LineUp           TermComands = "\033[F"
	OwnTextColour    TermComands = "\033[38;5;46;48;5;0m"
	OtherTextColour  TermComands = "\033[38;5;45;48;5;0m"
	ServerTextColour TermComands = "\033[38;5;214;48;5;0m"
	ResetColour      TermComands = "\033[39m\033[49m"
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
		fmt.Printf("%s<%s> ", ResetColour, UI.username.Uuid)
		msg := <-UI.reciveBuffer

		if msg.User.Uuid == UI.username.Uuid {
			fmt.Printf("%s%s", LineUp, ClearLine)
		}
		fmt.Printf("%s", ClearLine)
		// Display logical timestamp alongside physical timestamp
		switch msg.User.Uuid {
		case UI.username.Uuid:
			printMsg(msg, OwnTextColour)
		case "System":
			printMsg(msg, ServerTextColour)
		default:
			printMsg(msg, OtherTextColour)
		}
	}
}

func printMsg(msg *chitchat.Msg, colour TermComands) {
	outTime := time.Unix(msg.UnixTime, 0).Format(time.DateTime)
	fmt.Printf("%s<%s @ %s (L:%d)> %s%s\n", colour, msg.User.Uuid, outTime, msg.LogicalTime, msg.Message, ResetColour)
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
		fmt.Printf("<%s> ", UI.username.Uuid)
	}
}
