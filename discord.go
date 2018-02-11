package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"mvdan.cc/xurls"
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself, blacklisted members, channels it's not listening on
	if m.Author.Bot == true || blacklisted(m.Author.ID) == true || listenon(m.ChannelID) == false {
		writeLog("debug", "Author: "+m.Author.ID, nil)
		writeLog("debug", "Bot: "+strconv.FormatBool(m.Author.Bot), nil)
		writeLog("debug", "Blacklisted: "+strconv.FormatBool(blacklisted(m.Author.ID)), nil)
		writeLog("debug", "Channel: "+strconv.FormatBool(listening(m.ChannelID)), nil)
		writeLog("debug", "Message caught", nil)
		return
	}

	//
	// Message Handling
	//

	writeLog("debug", m.Content+"\n", nil)

	// Set input
	input := strings.ToLower(m.Content)

	// Reset response every message
	response = ""

	if strings.HasPrefix(input, viper.GetString("prefix")) == false {
		// If the prefix is not present
		if strings.Contains(input, ".png") == true || strings.Contains(input, ".jpg") {
			remoteURL := xurls.Relaxed().FindString(m.Content)
			input = parseImage(remoteURL)
			writeLog("debug", "Contains link to image", nil)
		}
		if strings.Contains(input, "astebin") == true {
			remoteURL := xurls.Relaxed().FindString(input)
			input = parseBin(remoteURL)
			writeLog("debug", "Is a bin link", nil)
		}
		response = parseChat(input)
	} else if strings.HasPrefix(input, viper.GetString("prefix")) == true {
		// If the prefix is present
		if strings.Contains(input, "ggl") == true {
			writeLog("debug", "Googling for user.", nil)
			response = "<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(input, ".ggl "), " ", "+", -1) + ">"
		} else {
			response = parseCommand(input)
		}
		if response == "" {
			return
		}
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		writeLog("debug", "Cleared command message.", nil)

	} else {
		response = "That's not a recognized command."
	}

	writeLog("debug", "Job's done", nil)

	if response == "" {
		return
	}
	writeLog("debug", "Message Sent"+response+"\n", nil)
	s.ChannelMessageSend(m.ChannelID, response)
}

func startDiscordConnection() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + "")

	writeLog("debug", "using "+""+" for discord", nil)

	if err != nil {
		writeLog("fatal", "error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		writeLog("fatal", "error obtaining Discord account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		writeLog("fatal", "error opening connection,", err)
		return
	}
}
