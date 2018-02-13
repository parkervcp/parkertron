package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"mvdan.cc/xurls"
)

func channelFilter(req string) bool {
	if getDiscordConfigBool("discord.channels.filter") == true {
		if strings.Contains(getDiscordChannels(), req) {
			return true
		}
	}
	return false
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself, blacklisted members, channels it's not listening on
	if m.Author.Bot == true || strings.Contains(getDiscordGroupMembers("blacklist"), m.Author.ID) == true || channelFilter(m.ChannelID) == false {
		if m.Author.Bot == true {
			writeLog("debug", "User is a bot and being ignored.", nil)
		}
		if strings.Contains(getDiscordGroupMembers("blacklist"), m.Author.ID) == true {
			writeLog("debug", "User is blacklisted and being ignored.", nil)
		}
		if getDiscordConfigBool("discord.channels.filter") == true {
			writeLog("debug", "This channel is being filtered out and ignored.", nil)
		}
		writeLog("debug", "Message has been ignored.\n", nil)
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

	if strings.HasPrefix(input, getDiscordConfigString("prefix")) == false {
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
	} else if strings.HasPrefix(input, getDiscordConfigString("prefix")) == true {
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

	writeLog("debug", "using "+getDiscordConfigString("discord.token")+" for discord\n", nil)

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + getDiscordConfigString("discord.token"))

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
