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
	//BotID is
	BotID string
	//ShowConfig is
	ShowConfig string
)

func init() {

	flag.StringVar(&ShowConfig, "S", "", "Show Config")
	flag.Parse()
}

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
	Adm []string `json:"admin"`
	Blk []string `json:"blacklist"`
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

func getLines(a string) string {
	//Returns the lines from the commands file as a response.
	//Needs to actually send back the correct response.
	commands := getCommands()
	for _, p := range commands {
		if p.Cmd == a {
			return string(a)
		}
	}
	var badcmd string
	return badcmd
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

	// Set input
	input := m.Content

	// Ignore commands without prefix
	if strings.HasPrefix(input, getConfig("prefix")) == false {
		return
	}

	// Command with prefix gets ran
	if strings.HasPrefix(input, getConfig("prefix")) == true {
		input = strings.TrimPrefix(input, getConfig("prefix"))
		s.ChannelMessageSend(m.ChannelID, getLines(input))
	}

}
