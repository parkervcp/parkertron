package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/otiai10/gosseract"
	"mvdan.cc/xurls"

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

func parseImage(remoteURL string) string {
	log.Info("Reading from " + remoteURL)

	remote, e := http.Get(remoteURL)
	if e != nil {
		log.Fatal(e)
	}

	defer remote.Body.Close()
	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := remoteURL[lastBin+1:]

	log.Info("Filename is " + fileName)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()
	log.Info("Image File Pulled and saved to /tmp/" + fileName)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("/tmp/" + fileName)
	text, err := client.Text()
	if err != nil {
		log.Fatal(err.Error())
	}

	return text
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

	// If the prefix is not present
	if hasPrefix(input) == false {
		if strings.Contains(strings.ToLower(input), ".png") == true {
			remoteURL := xurls.Relaxed().FindString(input)
			input = parseImage(remoteURL)
		}
		for _, p := range commands {
			if strings.Contains(strings.ToLower(input), p.Cmd) {
				if p.Typ == "listen" {
					for _, line := range p.Lns {
						response = response + "\n" + line
					}
				}
			}
		}
	} else if hasPrefix(input) == true {
		log.Info("Cleared command message.")
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		//Trim prefix from command

		if strings.Contains(input, "ggl") {
			response = "<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(input, "ggl "), " ", "+", -1) + ">"
			return
		}
		//Search command file for command and prep response
		for _, p := range commands {
			if p.Cmd == strings.ToLower(strings.TrimPrefix(input, getConfig("prefix"))) {
				if p.Typ == "chat" {
					for _, line := range p.Lns {
						response = response + "\n" + line
					}
				} else if p.Typ == "group" {
					if getAdmin(m.Author.ID) == false {
					}
					s.ChannelMessageSend(m.ChannelID, "You're an admin")
				}
			}
		}
	} else {
		response = "That's not a recognized command."
	}

	if response == "" {
		return
	} else {
		log.Info("Message Sent" + "\n" + response)
		s.ChannelMessageSend(m.ChannelID, response)
	}
}
