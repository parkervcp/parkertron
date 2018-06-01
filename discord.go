package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	//BotSession is the DiscordSession
	dg *discordgo.Session
)

func channelFilter(req string) bool {
	if getDiscordConfigBool("channels.filter") == true {
		if strings.Contains(getDiscordChannels(), req) {
			return true
		}
		return false

	}
	return true
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

	input := m.Content

	if m.Author.ID == m.ChannelID {
		writeLog("debug", "A DM was received", nil)
		sendDiscordMessage(m.ChannelID, "I currently don't respond to DMs")
		return
	}

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == dg.State.User.ID {
			writeLog("debug", "The bot was mentioned\n", nil)
			if strings.Replace(input, "<@"+dg.State.User.ID+">", "", -1) == "" {
				sendDiscordMessage(m.ChannelID, "I was mentioned. How can I help?")
				return
			}
		}
	}

	//
	// Message Handling
	//
	if input != "" {
		writeLog("debug", "Message Content: "+input+"\n", nil)

		// If the string doesn't have the prefix parse as text, if it does parse as a command.
		if strings.HasPrefix(input, getDiscordConfigString("prefix")) == false {
			parseKeyword("discord", m.ChannelID, input)

		} else if strings.HasPrefix(input, getDiscordConfigString("prefix")) == true {
			input = strings.TrimPrefix(input, getDiscordConfigString("prefix"))
			parseCommand("discord", m.ChannelID, input)

			s.ChannelMessageDelete(m.ChannelID, m.ID)
			writeLog("debug", "Cleared command message. \n", nil)
		}
		return
	}
}

func sendDiscordMessage(ChannelID string, response string) {
	response = strings.Replace(response, "&prefix&", getDiscordConfigString("prefix"), -1)

	writeLog("debug", "ChannelID "+ChannelID+" \n Discord Message Sent: \n"+response+"\n", nil)
	dg.ChannelMessageSend(ChannelID, response)
}

func startDiscordConnection() {
	//Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + getDiscordConfigString("token"))

	if err != nil {
		writeLog("fatal", "error creating Discord session,", err)
		return
	}

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
