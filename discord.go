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
		if getDiscordKOMChannel(req) {
			return true
		}
		return false

	}
	return true
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
			for _, ment := range m.Mentions {
				if ment.ID == dg.State.User.ID {
					writeLog("debug", "The bot was mentioned\n", nil)
					sendDiscordMessage(m.ChannelID, getDiscordConfigString("mention.wrong_channel"))
				}
			}
		}
		writeLog("debug", "Message has been ignored.\n", nil)
		return
	}

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == dg.State.User.ID {
			writeLog("debug", "The bot was mentioned\n", nil)
			sendDiscordMessage(m.ChannelID, getDiscordConfigString("mention.response"))
			if strings.Replace(input, "<@"+dg.State.User.ID+">", "", -1) == "" {
				sendDiscordMessage(m.ChannelID, getDiscordConfigString("mention.empty"))
			}
		}
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		writeLog("fatal", "", err)
		return
	}

	if strings.Contains(getDiscordGroupMembers("admin"), m.Author.ID) || strings.Contains(getDiscordGroupMembers("mods"), m.Author.ID) {
		writeLog("debug", "User is in an admin group", nil)
	} else {
		// Listen only channel filter (no parsing)
		if getDiscordKOMChannel(m.ChannelID) {
			writeLog("debug", "Message is not being parsed but listened to.", nil)
			// Check if a group is mentioned in message
			for _, ment := range m.MentionRoles {
				writeLog("debug", "Group "+ment+" was Mentioned", nil)
				if strings.Contains(getDiscordKOMID(m.ChannelID+".group"), ment) {
					writeLog("info", "Sending message to channel", nil)
					sendDiscordMessage(m.ChannelID, "Be gone with you <@"+m.Author.ID+">")
					writeLog("info", "Sending message to user", nil)
					sendDiscordDirectMessage(m.Author.ID, getDiscordKOMID(m.ChannelID+".reason"))
					kickDiscordUser(channel.GuildID, m.Author.ID, getDiscordKOMID(m.ChannelID+".reason"))
				}
			}
			return
		}
	}

	// Respond on DM's
	// TODO: Make the response customizable
	if channel.Type == 1 {
		writeLog("debug", "This was a DM", nil)
		sendDiscordMessage(m.ChannelID, "Thank you for messaging me, but I only offer support in the main chat.")
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
			parseCommand("discord", m.ChannelID, m.Author.ID, input)
			// remove previous commands if
			if getDiscordConfigBool("command.remove") && getCommandStatus(input) {
				writeLog("debug", "Cleared command message. \n", nil)
				deleteDiscordMessage(channel.ID, m.ID)
			}
		}
		return
	}
}

func sendDiscordMessage(ChannelID string, response string) {
	response = strings.Replace(response, "&prefix&", getDiscordConfigString("prefix"), -1)

	writeLog("debug", "ChannelID "+ChannelID+" \n Discord Message Sent: \n"+response+"\n", nil)
	dg.ChannelMessageSend(ChannelID, response)
}

func sendDiscordDirectMessage(userID string, response string) {
	channel, err := dg.UserChannelCreate(userID)
	if err != nil {
		writeLog("fatal", "error creating direct message channel,", err)
		return
	}
	sendDiscordMessage(channel.ID, response)
}

func deleteDiscordMessage(ChannelID string, MessageID string) {
	dg.ChannelMessageDelete(ChannelID, MessageID)
}

func kickDiscordUser(guild string, user string, reason string) {
	writeLog("debug", "Guild: "+guild+"\nUser: "+user+"\nreason: "+reason, nil)
	dg.GuildMemberDeleteWithReason(guild, user, reason)
	writeLog("info", "User has been kicked", nil)
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
