package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
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
		log.Error(err.Error())
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
	// Ignore all messages created by the bot itself, blacklisted members, channels it's not listening on
	if m.Author.Bot == true || blacklisted(m.Author.ID) == true || listenon(m.ChannelID) == false {
		log.Debug(m.Author.ID)
		log.Debug(m.Author.Bot)
		log.Debug(blacklisted(m.Author.ID))
		log.Debug(listening(m.ChannelID))
		log.Debug("Message caught")
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
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		//Trim prefix from command
		input = strings.ToLower(strings.TrimPrefix(input, getConfig("prefix")))

		if strings.Contains(input, "ggl") {
			response = "<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(input, "ggl "), " ", "+", -1) + ">"
			log.Info("Message Sent" + response)
			s.ChannelMessageSend(m.ChannelID, response)
		} else {
			//Search command file for command and prep response
			for _, p := range commands {
				if p.Cmd == input {
					if p.Typ == "chat" {
						for _, line := range p.Lns {
							response = response + "\n" + line
							log.Info("Message Sent" + response)
							s.ChannelMessageSend(m.ChannelID, response)
						}
					} else if p.Typ == "group" {
						if getAdmin(m.Author.ID) == false {
							return
						}
						s.ChannelMessageSend(m.ChannelID, "You're an admin")
					}
				}
			}
		}
	} else if hasPrefix(input) == false {
		for _, p := range commands {
			if strings.Contains(input, p.Cmd) {
				if p.Typ == "listen" {
					for _, line := range p.Lns {
						response = response + "\n" + line
						log.Info("Message Sent" + response)
						s.ChannelMessageSend(m.ChannelID, response)
					}
				}
			}
		}
	} else {
		response = "That's not a recognized command."
	}
}
