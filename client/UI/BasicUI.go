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

const ClearLine = "\033[2K\r"

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
		fmt.Printf("<%s> ", UI.username.Uuid)
		msg := <-UI.reciveBuffer

		if msg.User.Uuid == UI.username.Uuid {
			fmt.Printf("\033[F%s", ClearLine)
		}
		outTime := time.Unix(msg.UnixTime, 0).Format(time.DateTime)
		fmt.Printf(ClearLine)
		// Display logical timestamp alongside physical timestamp
		fmt.Printf("<%s @ %s (L:%d)> %s\n", msg.User.Uuid, outTime, msg.LogicalTime, msg.Message)
	}
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

		commands := strings.SplitN(input, " ", 2)
		if len(commands) < 2 {
			commands = append(commands, "")
		}
		commands[1] = strings.TrimSpace(commands[1])
		commands = append(commands, "") //make shure index is not out of bounds
		commands = append(commands, "")

		if commands[0] == "Disconnect" {
			UI.mc.Disconenct()
		} else if commands[0] == "say" && len(commands[1]) != 0 {
			UI.sendBuffer <- commands[1]
			fmt.Printf("<%s> ", UI.username.Uuid)
		} else {
			fmt.Println(UI.helpMessage)
			fmt.Print("<&s> ", UI.username.Uuid)
		}
	}
}

