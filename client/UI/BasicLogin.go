package ui

import (
	"fmt"
	"log"

	termcommands "github.com/Mojjedrengen/ChitChat/client/Termcommands"
	chitchat "github.com/Mojjedrengen/ChitChat/grpc"
	"github.com/google/uuid"
)

func BasicLogin() *chitchat.User {
	fmt.Printf("%s%s%s", termcommands.Save, termcommands.SaveCursorSCO, termcommands.Clear)
	fmt.Println("Please enter your username:")
	var username string
	_, err := fmt.Scanln(&username)
	if err != nil {
		log.Fatalf("failed to get username: %v", err)
	}
	user := &chitchat.User{
		Uuid: uuid.NewString(),
		Name: username,
	}
	fmt.Printf("%s", termcommands.Clear)
	return user
}
