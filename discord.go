package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	//BotID is the Discord Bot ID
	BotID string
)

func channelFilter(req string) bool {
	if getDiscordConfigBool("channels.filter") == true {
		if strings.Contains(getDiscordChannels(), req) {
			return true
		}
	}
	return false
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// If the owner is making a message always parse
	// Ignore all messages created by the bot itself, blacklisted members, channels it's not listening on, with debug messaging.
	if m.Author.Bot == true || strings.Contains(getDiscordGroupMembers("blacklist"), m.Author.ID) == true || channelFilter(m.ChannelID) == false {
		if m.Author.Bot == true {
			writeLog("debug", "User is a bot and being ignored.", nil)
		}
		if strings.Contains(getDiscordGroupMembers("blacklist"), m.Author.ID) == true {
			writeLog("debug", "User is blacklisted and being ignored.", nil)
		}
		if channelFilter(m.ChannelID) == false {
			writeLog("debug", "This channel is being filtered out and ignored.", nil)
		}
		writeLog("debug", "Message has been ignored.\n", nil)
		return
	}

	//
	// Message Handling
	//

	writeLog("debug", "Message Content: "+m.Content+"\n", nil)

	// Reset response every message
	response = ""

	if strings.HasPrefix(m.Content, getDiscordConfigString("prefix")) == false {
		response = parseKeyword(m.Content)

	} else if strings.HasPrefix(m.Content, getDiscordConfigString("prefix")) == true {
		trimmed := strings.TrimPrefix(m.Content, getDiscordConfigString("prefix"))
		response = parseCommand(trimmed)

		if response == "" {
			return
		}

		s.ChannelMessageDelete(m.ChannelID, m.ID)
		writeLog("debug", "Cleared command message. \n", nil)

	} else {
		response = "That's not a recognized command."
	}

	if response == "" {
		return
	}

	writeLog("debug", "Message Sent: "+response+"\n", nil)
	s.ChannelMessageSend(m.ChannelID, response)
}

func startDiscordConnection() {
	//Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + getDiscordConfigString("token"))

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
