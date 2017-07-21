package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	nextSend = time.Now()
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

func hasPrefix(a string) bool {
	return strings.HasPrefix(a, getConfig("prefix"))
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	if getChannelStat() == true {
		if listening(m.ChannelID) == false {
			return
		}
	}

	// Ignore all users on blacklist
	if blacklisted(m.Author.ID) == true {
		return
	}

	//	Ignore if Cooldown is still active
	if time.Now().Before(nextSend) {
		return
	}

	//
	// Message Handling
	//

	commands := getCommands()

	// Set input
	input := m.Content

	// Reset response every message
	response = ""

	// If the prefix is present
	if hasPrefix(input) == true {

		//Trim prefix from command
		input = strings.TrimPrefix(input, getConfig("prefix"))

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
	} else if hasPrefix(input) == false {
		for _, p := range commands {
			if strings.Contains(input, p.Cmd) {
				if p.Typ == "listen" {
					for _, line := range p.Lns {
						response = response + "\n" + line
					}
				}
			}
		}
	}

	//Send response
	s.ChannelMessageSend(m.ChannelID, response)
	nextSend = time.Now().Add(time.Second * time.Duration(getCooldown()))
}
