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
	// Reset response every message
	response = ""

	input := m.Content

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

	if m.Author.ID == m.ChannelID {
		writeLog("debug", "This was a DM", nil)
	}

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == BotID {
			writeLog("debug", "The bot was mentioned\n", nil)
			if strings.Replace(input, "<@"+BotID+">", "", -1) == "" {
				input = ""
				response = "I was mentioned. How can I help?"
			}
		}
	}

	//
	// Message Handling
	//
	if input != "" {
		writeLog("debug", "Message Content: "+input+"\n", nil)

		if strings.HasPrefix(input, getDiscordConfigString("prefix")) == false {
			response = parseKeyword(input)

		} else if strings.HasPrefix(input, getDiscordConfigString("prefix")) == true {
			trimmed := strings.TrimPrefix(input, getDiscordConfigString("prefix"))
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
	}
	response = strings.Replace(response, "&prefix&", getDiscordConfigString("prefix"), -1)

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

	writeLog("info", "Discord service connected\n", nil)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		writeLog("fatal", "error opening connection,", err)
		return
	}
	writeLog("info", "Discord service started\n", nil)

	ServStat <- "discord_online"
}
