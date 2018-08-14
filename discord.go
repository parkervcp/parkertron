package main

import (
	"fmt"
	"strconv"
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
	dpack := DataPackage{}
	authed := false

	dpack.Service = "discord"

	// Ignore all messages created by bots
	if m.Author.Bot {
		debug("User is a bot and being ignored.")
		return
	}

	// get channel information
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		fatal("Channel error", err)
		return
	}

	// Respond on DM's
	// TODO: Make the response customizable
	if channel.Type == 1 {
		debug("This was a DM")
		sendDiscordMessage(dpack, getDiscordConfigString("direct.response"))
		return
	}

	// get guild info
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		fatal("Guild error", err)
		return
	}

	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// quick referrence for information
	dpack.Message = m.Content
	dpack.MessageID = m.ID
	dpack.AuthorID = m.Author.ID
	dpack.AuthorName = m.Author.Username
	dpack.BotID = bot.ID
	dpack.ChannelID = channel.ID
	dpack.GuildID = guild.ID

	// get group status. If perms are set and group name. These are note weighted yet.
	dpack.Perms, dpack.Group = discordPermissioncheck(dpack.AuthorID)

	// setting server owner default to admin perms
	if dpack.AuthorID == guild.OwnerID {
		dpack.Perms = true
		dpack.Group = "admin"
		authed = true
	}

	// debug messaging
	if dpack.Perms {
		debug("author has perms and is in the group: " + dpack.Group)
		if dpack.Group == "admin" || dpack.Group == "mod" {
			authed = true
		}
	}

	// kick for mentioning a group in a specific channel.
	if getDiscordKOMChannel(dpack.ChannelID) {
		if !dpack.Perms {
			debug("Message is not being parsed but listened to.")
			// Check if a group is mentioned in message
			for _, ment := range m.MentionRoles {
				debug("Group " + ment + " was Mentioned")
				if strings.Contains(getDiscordKOMID(dpack.ChannelID+".group"), ment) {
					debug("Sending message to channel")
					sendDiscordMessage(dpack, getDiscordKOMMessage(dpack.ChannelID))
					debug("Sending message to user")
					sendDiscordDirectMessage(dpack, getDiscordKOMID(dpack.ChannelID+".reason"))
					kickDiscordUser(guild.ID, dpack.AuthorID, dpack.AuthorName, getDiscordKOMID(dpack.ChannelID+".reason"), dpack.BotID)
				}
			}
		}
		return
	}

	// ignore blacklisted users
	if strings.Contains(getDiscordBlacklist(), dpack.AuthorID) == true {
		debug("User is blacklisted and being ignored.")
	}

	// making a string array for attached images on messages.
	for _, y := range m.Attachments {
		debug(y.ProxyURL)
		dpack.Attached = append(dpack.Attached, y.ProxyURL)
	}

	// Always parse owner and group commands. Keyswords in matched channels.
	if !authed {
		// Ignore all messages created by blacklisted members, channels it's not listening on, with debug messaging.
		if !discordChannelFilter(dpack.ChannelID) {
			debug("This channel is being filtered out and ignored.")
			for _, ment := range m.Mentions {
				if ment.ID == dg.State.User.ID {
					debug("The bot was mentioned")
					sendDiscordMessage(dpack, getDiscordConfigString("mention.wrong_channel"))
				}
			}
			debug("Message has been ignored.")
			return
		}
	}

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == dg.State.User.ID {
			debug("The bot was mentioned")
			sendDiscordMessage(dpack, getDiscordConfigString("mention.response"))
			if strings.Replace(message, "<@"+dg.State.User.ID+">", "", -1) == "" {
				sendDiscordMessage(dpack, getDiscordConfigString("mention.empty"))
			}
		}
	}

	//
	// Message Handling
	//
	if message != "" {
		debug("Message Content: " + message)
		discordMessageHandler(dpack)
		return
	}
}

func sendDiscordMessage(dpack DataPackage, response string) {
	response = strings.Replace(response, "&prefix&", getDiscordConfigString("prefix"), -1)

	if strings.Contains(response, "&react&") {
		response = strings.Replace(response, "&react&", "", -1)
		//discordReaction()
	}

	if strings.Contains(response, "&user&") {
		response = strings.Replace(response, "&user&", "", -1)
		//discordReaction()
	}

	superdebug("ChannelID " + dpack.ChannelID + " \n Discord Message Sent: \n" + response)
	dg.ChannelMessageSend(dpack.ChannelID, response)
}

func deleteDiscordMessage(dpack DataPackage) {
	dg.ChannelMessageDelete(dpack.ChannelID, dpack.MessageID)

	embed := &discordgo.MessageEmbed{
		Title: "Message was deleted",
		Color: 0xf39c12,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "MessageID",
				Value:  dpack.MessageID,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Message Content",
				Value:  dpack.Message,
				Inline: true,
			},
		},
	}

	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	superdebug("message was deleted.")
}

func sendDiscordReaction(channelID string, messageID string, emojiID string, userID string, job string) {
	if job == "add" {
		dg.MessageReactionAdd(channelID, messageID, emojiID)
	}
	if job == "remove" {
		dg.MessageReactionRemove(channelID, messageID, emojiID, userID)
	}
}

func sendDiscordDirectMessage(dpack DataPackage, response string) {
	channel, err := dg.UserChannelCreate(dpack.AuthorID)
	dpack.ChannelID = channel.ID
	if err != nil {
		fatal("error creating direct message channel,", err)
		return
	}
	sendDiscordMessage(dpack, response)
}

func kickDiscordUser(guild string, user string, username string, reason string, authorname string) {
	dg.GuildMemberDeleteWithReason(guild, user, reason)

	embed := &discordgo.MessageEmbed{
		Title: "User has been kicked",
		Color: 0xf39c12,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "User",
				Value:  username,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "By",
				Value:  authorname,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Reason",
				Value:  reason,
				Inline: true,
			},
		},
	}

	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	info("User " + authorname + " has been kicked from " + guild + " for " + reason)
}

func banDiscordUser(guild string, user string, username string, reason string, days int, authorname string) {
	dg.GuildBanCreateWithReason(guild, user, reason, days)

	embed := &discordgo.MessageEmbed{
		Title: "User has been banned for " + strconv.Itoa(days) + " days",
		Color: 0xc0392b,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "User",
				Value:  username,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "By",
				Value:  authorname,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Reason",
				Value:  reason,
				Inline: true,
			},
		},
	}

	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	info("User " + authorname + " has been kicked from " + guild + " for " + reason)
}

func sendDiscordEmbed(channelID string, embed *discordgo.MessageEmbed) {
	_, err := dg.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		fatal("Embed send error", err)
		return
	}
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
