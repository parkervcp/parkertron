package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	//BotSession is the DiscordSession
	dg *discordgo.Session
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.Bot {
		debug("User is a bot and being ignored.")
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fatal("Channel error", err)
		return
	}

	// Respond on DM's
	// TODO: Make the response customizable
	if channel.Type == 1 {
		debug("This was a DM")
		sendDiscordMessage(channel.ID, getDiscordConfigString("direct.response"))
		return
	}

	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fatal("Guild error", err)
		return
	}

	message := m.Content
	messageID := m.ID
	author := m.Author.ID
	channelID := channel.ID
	attachments := m.Attachments

	perms, group := discordPermissioncheck(author)

	if author == guild.OwnerID {
		perms = true
		group = "admin"
	}

	if perms {
		debug("author has perms and is in the group: " + group)
	}

	var attached []string

	for _, y := range attachments {
		debug(y.ProxyURL)
		attached = append(attached, y.ProxyURL)
		discordAttachmentHandler(attached, channelID)
	}

	// If the owner is making a message always parse
	// Ignore all messages created by blacklisted members, channels it's not listening on, with debug messaging.
	if strings.Contains(getDiscordBlacklist(), author) || !discordChannelFilter(channelID) {
		if strings.Contains(getDiscordBlacklist(), author) == true {
			debug("User is blacklisted and being ignored.")
		}
		if discordChannelFilter(channelID) == false {
			debug("This channel is being filtered out and ignored.")
			for _, ment := range m.Mentions {
				if ment.ID == dg.State.User.ID {
					debug("The bot was mentioned")
					sendDiscordMessage(channelID, getDiscordConfigString("mention.wrong_channel"))
				}
			}
		}
		debug("Message has been ignored.")
		return
	}

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == dg.State.User.ID {
			debug("The bot was mentioned")
			sendDiscordMessage(channelID, getDiscordConfigString("mention.response"))
			if strings.Replace(message, "<@"+dg.State.User.ID+">", "", -1) == "" {
				sendDiscordMessage(channelID, getDiscordConfigString("mention.empty"))
			}
		}
	}

	if getDiscordKOMChannel(channelID) {
		if author != guild.OwnerID || !perms {
			debug("Message is not being parsed but listened to.")
			// Check if a group is mentioned in message
			for _, ment := range m.MentionRoles {
				debug("Group " + ment + " was Mentioned")
				if strings.Contains(getDiscordKOMID(channelID+".group"), ment) {
					debug("Sending message to channel")
					sendDiscordMessage(channelID, "Be gone with you <@"+author+">")
					debug("Sending message to user")
					sendDiscordDirectMessage(author, getDiscordKOMID(channelID+".reason"))
					kickDiscordUser(channel.GuildID, author, getDiscordKOMID(channelID+".reason"))
				}
			}
		}
		return
	}

	//
	// Message Handling
	//
	if message != "" {
		debug("Message Content: " + message)
		discordMessageHandler(message, channelID, messageID, author, perms, group)
		return
	}
}

func discordReaction(channelID string, messageID string, emojiID string, userID string, job string) {
	if job == "add" {
		dg.MessageReactionAdd(channelID, messageID, emojiID)
	}
	if job == "remove" {
		dg.MessageReactionRemove(channelID, messageID, emojiID, userID)
	}
}

func sendDiscordMessage(ChannelID string, response string) {
	response = strings.Replace(response, "&prefix&", getDiscordConfigString("prefix"), -1)

	superdebug("ChannelID " + ChannelID + " \n Discord Message Sent: \n" + response)
	dg.ChannelMessageSend(ChannelID, response)
}

func sendDiscordDirectMessage(userID string, response string) {
	channel, err := dg.UserChannelCreate(userID)
	if err != nil {
		fatal("error creating direct message channel,", err)
		return
	}
	sendDiscordMessage(channel.ID, response)
}

func deleteDiscordMessage(ChannelID string, MessageID string) {
	dg.ChannelMessageDelete(ChannelID, MessageID)
}

func kickDiscordUser(guild string, user string, reason string) {
	debug("Guild: " + guild + "\nUser: " + user + "\nreason: " + reason)
	dg.GuildMemberDeleteWithReason(guild, user, reason)
	debug("User has been kicked")
}

func startDiscordConnection() {
	//Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + getDiscordConfigString("token"))

	if err != nil {
		fatal("error creating Discord session,", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	debug("Discord service connected\n")

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fatal("error opening connection,", err)
		return
	}
	debug("Discord service started\n")

	ServStat <- "discord_online"
}
