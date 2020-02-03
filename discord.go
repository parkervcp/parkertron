package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

var (
	stopDiscord = make(map[string]chan string)

	discordGlobal discord

	discordLoad = make(chan string)
)

// This function will be called (due to AddHandler) when the bot receives
// the "ready" event from Discord.
func readyDiscord(dg *discordgo.Session, event *discordgo.Ready, game string) {
	// if there is an error setting the game log and return
	if err := dg.UpdateStatus(0, game); err != nil {
		Log.Fatalf("error setting game: %s", err)
		return
	}

	Log.Debugf("set game to: %s", game)
}

// This function will be called (due to AddHandler) every time a new
// message is created on any channel that the autenticated bot has access to.
func discordMessageHandler(dg *discordgo.Session, m *discordgo.MessageCreate, serverConfig *discordServer, DMResp responseArray) {
	// channel setting stuff
	var blacklistedUsers []string
	var channelCommands []command
	var channelKeywords []keyword
	var channelParsing parsing

	// internal use only
	var allChannels []string

	// data to send to discord
	var response []string
	var reaction []string

	// Ignore all messages created by bots (stops the bot uprising)
	if m.Author.Bot {
		Log.Debug("User is a bot and being ignored.")
		return
	}

	// load users that are in blacklisted groups
	for _, perms := range serverConfig.Permissions {
		if perms.Blacklisted {
			for _, user := range perms.Users {
				blacklistedUsers = append(blacklistedUsers, user)
			}
		}
	}

	// prep stuff for passing to the parser
	for _, group := range serverConfig.ChanGroups {
		for _, channel := range group.ChannelIDs {
			if m.ChannelID == channel {
				for _, command := range group.Commands {
					channelCommands = append(channelCommands, command)
				}
				for _, keyword := range group.Keywords {
					channelKeywords = append(channelKeywords, keyword)
				}
			}
		}
	}

	// get channel information
	channel, err := dg.State.Channel(m.ChannelID)
	if err != nil {
		Log.Fatal("Channel error", err)
		return
	}

	// if the channel is a DM
	if channel.Type == 1 {
		sendDiscordMessage(dg, m.ChannelID, m.Author.ID, serverConfig.Config.Prefix, DMResp.Response)
		sendDiscordReaction(dg, m.ChannelID, m.ID, DMResp.Reaction)
		return
	}

	// append all channels into the array
	for _, group := range serverConfig.ChanGroups {
		for _, channel := range group.ChannelIDs {
			allChannels = append(allChannels, channel)
		}
	}

	// if the channel isn't in a group drop the message
	for !strings.Contains(strings.Join(allChannels, ", "), m.ChannelID) {
		Log.Debugf("channel not found")
		return
	}

	// drop messages from blacklisted users
	if strings.Contains(strings.Join(blacklistedUsers, ", "), m.Author.ID) {
		Log.Debugf("user %s is blacklisted username is %s", m.Author.ID, m.Author.Username)
		return
	}

	var attachmentURLs []string
	for _, url := range m.Attachments {
		attachmentURLs = append(attachmentURLs, url.ProxyURL)
	}

	if strings.HasPrefix(m.Content, serverConfig.Config.Prefix) {
		response, reaction = parseCommand(strings.TrimPrefix(m.Content, serverConfig.Config.Prefix), dg.State.User.Username, channelCommands)
	} else {
		response, reaction = parseKeyword(m.Content, dg.State.User.Username, attachmentURLs, channelKeywords, channelParsing)
	}

	sendDiscordMessage(dg, m.ChannelID, m.Author.Username, serverConfig.Config.Prefix, response)
	sendDiscordReaction(dg, m.ChannelID, m.ID, reaction)
}

// kick a user and log it to a channel if configured
func kickDiscordUser(dg *discordgo.Session, guild, user, username, reason, authorname string) (err error) {
	if err = dg.GuildMemberDeleteWithReason(guild, user, reason); err != nil {
		return
	}

	// embed := &discordgo.MessageEmbed{
	// 	Title: "User has been kicked",
	// 	Color: 0xf39c12,
	// 	Fields: []*discordgo.MessageEmbedField{
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "User",
	// 			Value:  username,
	// 			Inline: true,
	// 		},
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "By",
	// 			Value:  authorname,
	// 			Inline: true,
	// 		},
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "Reason",
	// 			Value:  reason,
	// 			Inline: true,
	// 		},
	// 	},
	// }

	// TODO: Need to use new config for this
	// sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)

	return
}

// ban a user and log it to a channel if configured
func banDiscordUser(dg *discordgo.Session, guild, user, username, reason, authorname string, days int) (err error) {
	if err = dg.GuildBanCreateWithReason(guild, user, reason, days); err != nil {
		return
	}

	// embed := &discordgo.MessageEmbed{
	// 	Title: "User has been banned for " + strconv.Itoa(days) + " days",
	// 	Color: 0xc0392b,
	// 	Fields: []*discordgo.MessageEmbedField{
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "User",
	// 			Value:  username,
	// 			Inline: true,
	// 		},
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "By",
	// 			Value:  authorname,
	// 			Inline: true,
	// 		},
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "Reason",
	// 			Value:  reason,
	// 			Inline: true,
	// 		},
	// 	},
	// }

	// TODO: Need to use new config for embed audit to log to a webhook
	//	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)

	return
}

// clean up messages if configured to
func deleteDiscordMessage(dg *discordgo.Session, channelID, messageID, message string) (err error) {
	Log.Debugf("Removing message \n'%s'\n from %s", message, channelID)

	if err = dg.ChannelMessageDelete(channelID, messageID); err != nil {
		return
	}

	// embed := &discordgo.MessageEmbed{
	// 	Title: "Message was deleted",
	// 	Color: 0xf39c12,
	// 	Fields: []*discordgo.MessageEmbedField{
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "MessageID",
	// 			Value:  messageID,
	// 			Inline: true,
	// 		},
	// 		&discordgo.MessageEmbedField{
	// 			Name:   "Message Content",
	// 			Value:  message,
	// 			Inline: true,
	// 		},
	// 	},
	// }

	// TODO: Need to use new config for embed audit to log to a webhook
	// 	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Debug("message was deleted.")

	return
}

// send message handling
func sendDiscordMessage(dg *discordgo.Session, channelID, authorID, prefix string, responseArray []string) (err error) {
	// if there is no response to sen just return
	if len(responseArray) == 0 {
		return
	}

	response := strings.Join(responseArray, "\n")
	response = strings.Replace(response, "&user&", authorID, -1)
	response = strings.Replace(response, "&prefix&", prefix, -1)
	response = strings.Replace(response, "&react&", "", -1)

	// if there is an error return the error
	if _, err = dg.ChannelMessageSend(channelID, response); err != nil {
		return
	}

	return
}

// send a reaction to a message
func sendDiscordReaction(dg *discordgo.Session, channelID string, messageID string, reactionArray []string) (err error) {
	// if there is no reaction to sen just return
	if len(reactionArray) == 0 {
		return
	}

	for _, reaction := range reactionArray {
		Log.Debugf("sending \"%s\" as a reaction to message: %s", reaction, messageID)
		// if there is an error sending a message return it
		if err = dg.MessageReactionAdd(channelID, messageID, reaction); err != nil {
			return
		}
	}
	return
}

// send a message with an embed
func sendDiscordEmbed(dg *discordgo.Session, channelID string, embed *discordgo.MessageEmbed) error {
	// if there is an error sending the embed message
	if _, err := dg.ChannelMessageSendEmbed(channelID, embed); err != nil {
		Log.Fatal("Embed send error")
		return err
	}

	return nil
}

// service handling
// start all the bots
func startDiscordsBots() {
	Log.Infof("Starting IRC server connections\n")
	// range over the bots available to start
	for _, bot := range discordGlobal.Bots {
		Log.Infof("Connecting to %s\n", bot.BotName)

		// spin up a channel to tell the bot to shutdown later
		stopDiscord[bot.BotName] = make(chan string)

		// start the bot
		go startDiscordBotConnection(&bot)
		// wait on bot being able to start.
		<-discordLoad
	}

	Log.Debug("Discord service started\n")
	servStart <- "discord_online"
}

// when a shutdown is sent close out services properly
func stopDiscordBots() {
	Log.Infof("stopping discord connections")
	// loop through bots and send shutdowns
	for _, bot := range discordGlobal.Bots {
		Log.Infof("stopping %s", bot.BotName)
		stopDiscord[bot.BotName] <- ""

		<-stopDiscord[bot.BotName]
		Log.Infof("stopped %s", bot.BotName)
	}
	Log.Infof("discord connections stopped")
	// return shutdown signal on channel
	servStopped <- "discord_stopped"
}

// start connections to discord
func startDiscordBotConnection(discordConfig *discordBot) {
	// Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + discordConfig.Config.Token)
	if err != nil {
		Log.Errorf("error creating Discord session for %s: %v", discordConfig.BotName, err)
		return
	}

	// Register ready as a callback for the ready events
	// dg.AddHandler(readyDiscord)
	// Thank Stroom on the discordgopher discord for getting me this
	dg.AddHandler(func(dg *discordgo.Session, event *discordgo.Ready) {
		readyDiscord(dg, event, discordConfig.Config.Game)
	})

	// Register messageCreate as a callback for the messageCreate events.
	// dg.AddHandler(discordMessageHandler)
	// Thank Stroom on the discordgopher discord for getting me this
	for _, server := range discordConfig.Servers {
		dg.AddHandler(func(dg *discordgo.Session, event *discordgo.MessageCreate) {
			discordMessageHandler(dg, event, &server, discordConfig.Config.DMResp)
		})
	}

	Log.Debug("Discord service connected\n")

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		Log.Fatal("error opening connection,", err)
		return
	}

	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	Log.Debug("Invite the bot to your server with https://discordapp.com/oauth2/authorize?client_id=" + bot.ID + "&scope=bot")

	discordLoad <- ""

	<-stopDiscord[discordConfig.BotName]

	// properly send a shutdown to the discord server so the bot goes offline.
	dg.Close()

	// return the shutdown signal
	stopDiscord[discordConfig.BotName] <- ""
}
