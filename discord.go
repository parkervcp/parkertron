package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	Log "github.com/sirupsen/logrus"
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
		Log.Debug("User is a bot and being ignored.")
		return
	}

	// get channel information
	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		Log.Fatal("Channel error", err)
		return
	}

	// Respond on DM's
	// TODO: Make the response customizable
	if channel.Type == 1 {
		Log.Debug("This was a DM")
		dpack.Response = getDiscordConfigString("direct.response")
		sendDiscordMessage(dpack)
		return
	}

	// get guild info
	guild, err := s.Guild(channel.GuildID)
	if err != nil {
		Log.Fatal("Guild error", err)
		return
	}

	authorGuildInfo, err := dg.GuildMember(guild.ID, m.Author.ID)
	if err != nil {
		Log.Fatal("Author Guild Info error", err)
		return
	}

	//ignore messages from the bot
	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// quick referrence for information
	dpack.Message = m.Content
	dpack.MessageID = m.ID
	dpack.AuthorID = m.Author.ID
	dpack.AuthorName = m.Author.Username
	dpack.AuthorRoles = authorGuildInfo.Roles
	dpack.BotID = bot.ID
	dpack.ChannelID = channel.ID
	dpack.GuildID = guild.ID

	// get group status. If perms are set and group name. These are note weighted yet.
	dpack.Perms, dpack.Group = discordAuthorRolePermissionCheck(dpack.AuthorRoles)
	if !dpack.Perms {
		dpack.Perms, dpack.Group = discordAuthorUserPermissionCheck(dpack.AuthorID)
	}

	// setting server owner default to admin perms
	if dpack.AuthorID == guild.OwnerID {
		dpack.Perms = true
		dpack.Group = "admin"
		authed = true
	}

	// debug messaging
	if dpack.Perms {
		Log.Debug("User has perms and is in the group: " + dpack.Group)
		if dpack.Group == "admin" || dpack.Group == "mod" {
			authed = true
		}
	} else {
		Log.Debug("User has no perms")
	}

	// kick for mentioning a group in a specific channel.
	if getDiscordKOMChannel(dpack.ChannelID) {
		if !dpack.Perms {
			Log.Debug("Checking for Kick on Mention group")
			// Check if a group is mentioned in message
			for _, ment := range m.MentionRoles {
				Log.Debug("Group " + ment + " was Mentioned")
				if strings.Contains(getDiscordKOMID(dpack.ChannelID+".group"), ment) {
					dpack.Mention = dpack.AuthorID
					Log.Debug("Sending message to channel")
					dpack.Response = getDiscordKOMMessage(dpack.ChannelID)
					sendDiscordMessage(dpack)
					Log.Debug("Sending message to user")
					dpack.Response = getDiscordKOMID(dpack.ChannelID + ".reason")
					sendDiscordDirectMessage(dpack)
					kickDiscordUser(guild.ID, dpack.AuthorID, dpack.AuthorName, getDiscordKOMID(dpack.ChannelID+".reason"), dpack.BotID)
				}
			}
		}
	}

	Log.Debug("Passed KOM")

	// ignore blacklisted users
	if strings.Contains(getDiscordBlacklist(), dpack.AuthorID) {
		Log.Debug("User is blacklisted and being ignored.")
	}

	Log.Debug("Passed Blacklist")

	// making a string array for attached images on messages.
	for _, y := range m.Attachments {
		Log.Debug(y.ProxyURL)
		dpack.Attached = append(dpack.Attached, y.ProxyURL)
	}

	Log.Debug("Attachments grabbed")

	// Always parse owner and group commands. Keyswords in matched channels.
	if !authed {
		// Ignore all channels it's not listening on, with debug messaging.
		if !discordChannelFilter(dpack.ChannelID) {
			Log.Debug("This channel is being filtered out and ignored.")
			for _, ment := range m.Mentions {
				if ment.ID == dg.State.User.ID {
					Log.Debug("The bot was mentioned")
					dpack.Response = getDiscordConfigString("mention.wrong_channel")
					sendDiscordMessage(dpack)
				}
			}
			Log.Debug("Message has been ignored.")
			return
		}
	}

	Log.Debug("Passed channel filter")

	// Check if the bot is mentioned
	for _, ment := range m.Mentions {
		if ment.ID == dg.State.User.ID {
			Log.Debug("The bot was mentioned")
			dpack.Response = getDiscordConfigString("mention.response")
			sendDiscordMessage(dpack)
			if strings.Replace(message, "<@"+dg.State.User.ID+">", "", -1) == "" {
				dpack.Response = getDiscordConfigString("mention.empty")
				sendDiscordMessage(dpack)
			}
		}
	}

	Log.Debug("Passed bot mentions")

	//
	// Message Handling
	//
	if dpack.Message != "" || dpack.Attached != nil {
		Log.Debug("Message ID: " + dpack.MessageID + "\nMessage Content: " + dpack.Message)
		discordMessageHandler(dpack)
		return
	}
	// exists solely because it got here somehow...
	Log.Debug("Really...")
}

func sendDiscordMessage(dpack DataPackage) {
	if dpack.Response == "" {
		return
	}

	// parse for prefix in the response
	dpack.Response = strings.Replace(dpack.Response, "&prefix&", getDiscordConfigString("prefix"), -1)

	// parse for reactions in the response
	if strings.Contains(dpack.Response, "&react&") {
		dpack.Response = strings.Replace(dpack.Response, "&react&", "", -1)
		if dpack.MsgTye == "keyword" {
			dpack.Reaction = getKeywordReaction(dpack.Keyword)
		} else if dpack.MsgTye == "command" {
			dpack.Reaction = getCommandReaction(dpack.Command)
		}
		dpack.ReactAdd = true
	}

	//parse for user mentions in the response
	if strings.Contains(dpack.Response, "&user&") {
		if dpack.Mention == "" {
			dpack.Mention = dpack.AuthorID
		}
		dpack.Response = strings.Replace(dpack.Response, "&user&", "<@"+dpack.Mention+">", -1)
	}

	Log.Debug("ChannelID " + dpack.ChannelID + " \n Discord Message Sent: \n" + dpack.Response)

	//Send response to channel
	sent, err := dg.ChannelMessageSend(dpack.ChannelID, dpack.Response)
	if err != nil {
		Log.Fatal("error sending message", err)
		return
	}

	//send reaction to channel
	if dpack.ReactAdd {
		Log.Debug("Adding Reactions")
		sendDiscordReaction(sent.ChannelID, sent.ID, dpack)
	}

	// remove previous commands if discord.command.remove is true
	if getDiscordConfigBool("command.remove") {
		if getCommandStatus(dpack.Message) {
			deleteDiscordMessage(dpack)
			Log.Debug("Cleared command message.")
		}
		if strings.HasPrefix(dpack.Message, "list") || strings.HasPrefix(dpack.Message, "ggl") {
			deleteDiscordMessage(dpack)
			Log.Debug("Cleared command message.")
		}
	}
}

func deleteDiscordMessage(dpack DataPackage) {
	Log.Debug("Removing message: " + dpack.Message)
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

	if getDiscordConfigString("embed.audit") != "" {
		sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	}
	Log.Debug("message was deleted.")
}

func sendDiscordReaction(channelID string, messageID string, dpack DataPackage) {
	for _, reaction := range dpack.Reaction {
		Log.Debug("Adding reation \"" + reaction + "\" to message " + dpack.MessageID)
		dg.MessageReactionAdd(dpack.ChannelID, dpack.MessageID, reaction)
	}
}

func sendDiscordDirectMessage(dpack DataPackage) {
	channel, err := dg.UserChannelCreate(dpack.AuthorID)
	dpack.ChannelID = channel.ID
	if err != nil {
		Log.Fatal("error creating direct message channel.", err)
		return
	}
	sendDiscordMessage(dpack)
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

	if getDiscordConfigString("embed.audit") != "" {
		sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	}
	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)
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

	if getDiscordConfigString("embed.audit") != "" {
		sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)
	}
	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)
}

func sendDiscordEmbed(channelID string, embed *discordgo.MessageEmbed) {
	_, err := dg.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		Log.Fatal("Embed send error", err)
		return
	}
}

func startDiscordConnection() {
	//Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + getDiscordConfigString("token"))

	if err != nil {
		Log.Fatal("error creating Discord session,", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	Log.Debug("Discord service connected\n")

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		Log.Fatal("error opening connection,", err)
		return
	}
	Log.Debug("Discord service started\n")

	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	Log.Debug("Invite the bot to your server with https://discordapp.com/oauth2/authorize?client_id=" + bot.ID + "&scope=bot")

	ServStat <- "discord_online"
}
