package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

//Commands structure
type Commands struct {
	Cmd string   `json:"command"`
	Typ string   `json:"type"`
	Lns []string `json:"lines"`
}

func getCommands() []Commands {
	//Opens commands.json and returns values
	raw, err := ioutil.ReadFile("./commands.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var c []Commands
	json.Unmarshal(raw, &c)
	return c
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	//Ignore all users on blacklist
	if blacklisted(m.Author.ID) == true {
		return
	}

	//
	// Message Handling
	//

	// Set input
	input := m.Content

	// Ignore commands without prefix
	if strings.HasPrefix(input, getConfig("prefix")) == false {
		return
	}

	commands := getCommands()

	// Command with prefix gets ran
	if strings.HasPrefix(input, getConfig("prefix")) == true {
		//Reset response every message
		response = ""
		//Trim prefix from command
		input = strings.TrimPrefix(input, getConfig("prefix"))

		//drop prefix only commands
		if input == "" {
			return
		}

		//Search command file for command and prep response
		for _, p := range commands {
			if p.Cmd == input {
				if p.Typ == "chat" {
					for _, line := range p.Lns {
						response = response + "\n" + line
					}
				} else if p.Typ == "group" {
					if getAdmin(m.Author.ID) == false {
						return
					}
					s.ChannelMessageSend(m.ChannelID, "You're an admin")
				}
			}
		}
		//Send response
		s.ChannelMessageSend(m.ChannelID, response)
	}

}
