package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	//BotID is the Discord Bot ID
	BotID string
	//ShowConfig is part of the init process
	ShowConfig string
	//response is the bot response on the channel
	response string
)

//Config structure
type Config struct {
	Token  string `json:"token"`
	Client string `json:"client"`
	Owner  string `json:"owner"`
	Prefix string `json:"prefix"`
}

//Commands structure
type Commands struct {
	Cmd string   `json:"command"`
	Lns []string `json:"lines"`
}

//Perms structure
type Perms struct {
	Grp string   `json:"group"`
	UID []string `json:"uid"`
}

func init() {

	flag.StringVar(&ShowConfig, "S", "", "Show Config")
	flag.Parse()
}

func getConfig(a string) string {
	//Opens config.json and returns values
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error", err)
	}
	if a == "token" {
		var b = config.Token
		return b
	} else if a == "client" {
		var b = config.Client
		return b
	} else if a == "owner" {
		var b = config.Owner
		return b
	} else if a == "prefix" {
		var b = config.Prefix
		return b
	}
	var b = "error"
	return b
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

func getPerms() []Perms {
	//Opens commands.json and returns values
	raw, err := ioutil.ReadFile("./permissions.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var c []Perms
	json.Unmarshal(raw, &c)
	return c
}

func blacklisted(a string) bool {
	perms := getPerms()

	for _, p := range perms {
		if p.Grp == "blacklist" {
			for _, u := range p.UID {
				if u == a {
					return true
				}
			}
		}
	}
	return false
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + getConfig("token"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
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
		//Search command file for command and prep response
		for _, p := range commands {
			if p.Cmd == input {
				for _, line := range p.Lns {
					response = response + "\n" + line
				}
			}
		}
		//Send response
		s.ChannelMessageSend(m.ChannelID, response)
	}

}
