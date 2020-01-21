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

	stopDiscord = make(chan string)

	discordConfig discord
)

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func discordMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// exists solely because it got here somehow...
	Log.Debug("Really...")
}

func kickDiscordUser(guild, user, username, reason, authorname string) {
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

	fmt.Sprint(embed)

	// TODO: Need to use new config for this
	// sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)
}

func banDiscordUser(guild, user, username, reason, authorname string, days int) {
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

	fmt.Sprint(embed)

	// TODO: Need to use new config for embed audit to log to a webhook
	//	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Info("User " + authorname + " has been kicked from " + guild + " for " + reason)
}

// arbitrary message handling
func deleteDiscordMessage(channelID, messageID, message string) error {
	//Log.Debug("Removing message: " + dpack.Message)

	dg.ChannelMessageDelete(channelID, messageID)

	embed := &discordgo.MessageEmbed{
		Title: "Message was deleted",
		Color: 0xf39c12,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "MessageID",
				Value:  messageID,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Message Content",
				Value:  message,
				Inline: true,
			},
		},
	}

	fmt.Sprint(embed)

	// TODO: Need to use new config for embed audit to log to a webhook
	// 	sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Debug("message was deleted.")

	return nil
}

// send message handling
func sendDiscordMessage(s *discordgo.Session, channelID, authorID, prefix string, responseArray []string) error {
	response := strings.Join(responseArray, "\n")
	response = strings.Replace(response, "&user&", authorID, -1)
	response = strings.Replace(response, "&prefix&", prefix, -1)
	response = strings.Replace(response, "&react&", "", -1)

	_, err := s.ChannelMessageSend(channelID, response)
	if err != nil {
		return err
	}

	return nil
}

func sendDiscordReaction(s *discordgo.Session, channelID string, messageID string, reactionArray []string) {
	for _, reaction := range reactionArray {
		Log.Debugf("sending \"%s\" as a reaction to message: %s", reaction, messageID)
		err := s.MessageReactionAdd(channelID, messageID, reaction)
		if err != nil {
			Log.Errorf("There was an error sending the reaction. %s", err)
		}
	}
}

func sendDiscordEmbed(channelID string, embed *discordgo.MessageEmbed) error {
	_, err := dg.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		Log.Fatal("Embed send error")
		return err
	}

	return nil
}

// service handling
func startDiscordConnection() {
	loadConfigs(confDir)

	// Initializing Discord connection
	// Create a new Discord session using the provided bot token.
	dg, err = discordgo.New("Bot " + discordConfig.Token)

	if err != nil {
		Log.Fatal("error creating Discord session,", err)
		return
	}

	// Register ready as a callback for the ready events
	dg.AddHandler(readyDiscord)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(discordMessageHandler)

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

	servStat <- "discord_online"
	<-stopDiscord
	// properly send a shutdown to the discord server so the bot goes offline.
	dg.Close()
	stopDiscord <- ""
}

// when a shutdown is sent close out services properly
func stopDiscordConnection() {
	Log.Infof("stopping discord connection")
	stopDiscord <- ""
	<-stopDiscord
	Log.Infof("discord connection stopped")
	shutdown <- ""
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func readyDiscord(s *discordgo.Session, event *discordgo.Ready) {
	err := s.UpdateStatus(0, discordConfig.Token)
	if err != nil {
		Log.Fatalf("error setting game: %s", err)
		return
	}
	Log.Debugf("set game to: %s", discordConfig.Game)
}
