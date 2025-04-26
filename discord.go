package main

// Thank Stroom on the discordgopher discord for helping me with embedded functions in the handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"syscall"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/session"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
	"mvdan.cc/xurls/v2"
)

var (
	stopDiscordServer    = make(chan string)
	discordServerStopped = make(chan string)

	discordGlobal discordBase

	discordLoad = make(chan string)
)

// This function will be called (due to AddHandler) when the bot receives
// the "ready" event from Discord.
func readyDiscord(botSession *session.Session, game string) {
	if !discord.EmojiID(985546330271252530).IsValid() {
		Log.Debug("emoji is invalid")
	}

	activities := []discord.Activity{
		{
			Name:  "custom",
			Type:  4,
			State: game,
		},
	}

	status := &gateway.UpdatePresenceCommand{
		Activities: activities,
	}
	// if there is an error setting the game log and return
	if err := botSession.Gateway().Send(context.Background(), status); err != nil {
		Log.Fatalf("error setting game: %s", err)
		return
	}

	Log.Debugf("set game to: %s", game)
}

// This function will be called (due to AddHandler) every time a new
// message is created on any channel that the authenticated bot has access to.
func discordMessageHandler(botSession *session.Session, messageEvent *gateway.MessageCreateEvent, botName string) {
	Log.Debugf("bot is %s", botName)
	Log.Debugf("message '%s'", messageEvent.Content)

	// data to send to discord
	var response []string
	var reaction []string

	botUser, err := botSession.Me()
	if err != nil {
		fmt.Println("error obtaining account details,", err)
		syscall.Exit(2)
	}

	// Ignore all messages created by bots (stops the bot uprising)
	if messageEvent.Author.Bot {
		Log.Debug("User is a bot and being ignored.")
		return
	}

	// get channel information
	channel, err := botSession.Channel(messageEvent.ChannelID)
	if err != nil {
		Log.Fatal("Channel error ", err)
		return
	}

	message := messageEvent.Message

	botID := botUser.ID

	guildID := channel.GuildID.String()
	chanID := messageEvent.ChannelID.String()

	// if the channel type is a thread use the parent id for the config
	if channel.Type == 11 {
		chanID = channel.ParentID.String()
	}

	Log.Debugf("prefix: %s", getPrefix("discord", botName, guildID))

	// if the channel is a DM
	if channel.Type == 1 {
		_, dmResp := getMentions("discord", botName, guildID, "DirectMessage")
		if err := sendDiscordMessage(botSession, channel, messageEvent.Author, dmResp.Reaction, botName); err != nil {
			Log.Error(err)
		}

		if err := sendDiscordReaction(botSession, channel, message, dmResp.Reaction); err != nil {
			Log.Error(err)
		}

		return
	}

	// bot level configs for log reading
	maxLogs, logResponse, logReaction, allowIP := getBotParseConfig()

	//filter logic
	Log.Debug("filtering messages")
	if len(getFilter("discord", botName, guildID)) == 0 {
		Log.Debugf("no filtered terms found")
	} else {
		for _, filter := range getFilter("discord", botName, guildID) {
			if strings.Contains(messageEvent.Content, filter.Term) {
				Log.Infof("message was removed for containing %s", filter.Term)
				if err := deleteDiscordMessages(botSession, channel, []discord.MessageID{0: messageEvent.ID}, ""); err != nil {
					Log.Error(err)
				}

				if err := sendDiscordMessage(botSession, channel, messageEvent.Author, filter.Reason, botName); err != nil {
					Log.Error(err)
				}
				return
			} else {
				continue
			}
		}
	}

	// if the channel isn't in a group drop the message
	Log.Debugf("checking channels")
	if !contains(getChannels("discord", botName, guildID), chanID) {
		Log.Debugf("channel not found")
		return
	}

	Log.Debugf("checking blacklist")

	// drop messages from blacklisted users
	for _, user := range getBlacklist("discord", botName, guildID, chanID) {
		if user == messageEvent.Author.ID.String() {
			Log.Debugf("user %s is blacklisted username is %s", messageEvent.Author.ID.String(), messageEvent.Author.Username)
			return
		}
	}

	Log.Debugf("checking attachments")

	// for all attachment urls
	var attachmentURLs []string
	for _, url := range messageEvent.Attachments {
		attachmentURLs = append(attachmentURLs, url.Proxy)
	}

	// this was for debugging/testing only

	Log.Debugf("all attachments %s", attachmentURLs)
	Log.Debugf("all ignores %+v", getParsing("discord", botName, guildID, chanID).Paste.Ignore)

	Log.Debugf("checking for any urls in the message")
	var allURLS []string
	for _, url := range xurls.Relaxed().FindAllString(messageEvent.Content, -1) {
		Log.Debugf("checking on %s", url)
		// if the url is an ip filter it out
		if match, err := regexp.Match("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])", []byte(url)); err != nil {
			Log.Error(err)
		} else if match && !allowIP {
			Log.Debugf("adding %s to the list", url)
			continue
		}

		Log.Debugf("looking for ignored domains")
		if len(getParsing("discord", botName, guildID, chanID).Paste.Ignore) == 0 {
			Log.Debugf("appending %s to allURLS", url)
			allURLS = append(allURLS, url)
			Log.Debugf("no ignored domain found")
			continue
		} else {
			var ignored bool
			for _, ignoreURL := range getParsing("discord", botName, guildID, chanID).Paste.Ignore {
				Log.Debugf("url should be ignored: %t", strings.HasPrefix(url, ignoreURL.URL))
				if strings.HasPrefix(url, ignoreURL.URL) {
					ignored = true
					Log.Debugf("domain %s is being ignored.", ignoreURL.URL)
					break
				}
			}
			if ignored {
			} else {
				Log.Debugf("appending %s to allURLS", url)
				allURLS = append(allURLS, url)
			}
		}
	}

	// add all urls together
	Log.Debug("adding attachment URLS to allURLS")
	for i := 0; i < len(attachmentURLs); i++ {
		allURLS = append(allURLS, attachmentURLs[i])
	}

	// Log.Debug(allURLS)
	Log.Debugf("checking mentions")
	if len(messageEvent.Mentions) != 0 {
		ping, mention := getMentions("discord", botName, guildID, chanID)

		if messageEvent.Mentions[0].ID == botID && messageEvent.Content == fmt.Sprintf("<@%s>", botID.String()) {
			Log.Debugf("bot was pinged")
			response = ping.Response
			reaction = ping.Reaction
		} else {
			for _, mentioned := range messageEvent.Mentions {
				if mentioned.ID == botID {
					Log.Debugf("bot was mentioned")
					response = mention.Response
					reaction = mention.Reaction
				}
			}
		}
	} else {
		Log.Debugf("no mentions found")
	}

	if strings.HasPrefix(messageEvent.Content, getPrefix("discord", botName, guildID)) {
		// command
		response, reaction = parseCommand(strings.TrimPrefix(messageEvent.Content, getPrefix("discord", botName, guildID)), botName, getCommands("discord", botName, guildID, chanID))
		// if the flag for clearing commands is set and there is a response
		if getCommandClear("discord", botName, guildID) && len(response) > 0 {
			Log.Debugf("removing command message %s", messageEvent.ID.String())
			if err := deleteDiscordMessages(botSession, channel, []discord.MessageID{0: messageEvent.ID}, ""); err != nil {
				Log.Error(err)
			}
		}
	} else {
		// regex -- priority over keywords
		response, reaction = parseRegex(messageEvent.Content, botName, getRegexPatterns("discord", botName, guildID, chanID), getParsing("discord", botName, guildID, chanID))

		// keyword
		if response == nil {
			response, reaction = parseKeyword(messageEvent.Content, botName, getKeywords("discord", botName, guildID, chanID), getParsing("discord", botName, guildID, chanID))
		}
	}

	if len(getParsing("discord", botName, guildID, chanID).Image.FileTypes) == 0 && len(getParsing("discord", botName, guildID, chanID).Paste.Sites) == 0 {
		Log.Debugf("no parsing configs found")
	} else {
		Log.Debugf("allURLS: %s", allURLS)
		Log.Debugf("allURLS count: %d", len(allURLS))

		// if we have too many logs ignore it.
		if len(allURLS) == 0 {
			Log.Debugf("no URLs to read")
		} else if len(allURLS) > maxLogs {
			Log.Debug("too many logs or screenshots to try and read.")
			if err := sendDiscordMessage(botSession, channel, messageEvent.Author, logResponse, botName); err != nil {
				Log.Error(err)
			}
			if err := sendDiscordReaction(botSession, channel, message, logReaction); err != nil {
				Log.Error(err)
			}
			return
		} else {
			Log.Debugf("reading logs")
			if err := sendDiscordReaction(botSession, channel, message, []string{"👀"}); err != nil {
				Log.Error(err)
			}

			// get parsed content for each url/attachment
			Log.Debugf("reading all attachments and logs")
			allParsed := make(map[string]string)
			for _, url := range allURLS {
				allParsed[url] = parseURL(url, getParsing("discord", botName, guildID, chanID))
			}

			//parse logs and append to current response.
			for _, url := range allURLS {
				Log.Debugf("passing %s to keyword parser", url)
				urlResponse, _ := parseKeyword(allParsed[url], botName, getKeywords("discord", botName, guildID, chanID), getParsing("discord", botName, guildID, chanID))
				Log.Debugf("response length = %d", len(urlResponse))
				if len(urlResponse) == 1 && urlResponse[0] == "" || len(urlResponse) == 0 {

				} else {
					response = append(response, fmt.Sprintf("I have found the following for: <%s>", url))
					for _, singleLine := range urlResponse {
						response = append(response, singleLine)
					}
				}
			}
		}
	}

	// send response to channel
	Log.Debugf("sending response %s to %s", response, chanID)
	if err := sendDiscordMessage(botSession, channel, messageEvent.Author, response, botName); err != nil {
		Log.Error(err)
	}

	// send reaction to channel
	Log.Debugf("sending reaction %s", reaction)
	if err := sendDiscordReaction(botSession, channel, message, reaction); err != nil {
		Log.Error(err)
	}
}

// This function will be called (due to AddHandler) every time a new
// thread is created on any channel that the authenticated bot has access to.
func discordNewThreadHandler(botSession *session.Session, m *gateway.ThreadCreateEvent, botName string) {

}

// This function will be called (due to AddHandler) every time a new
// thread is created on any channel that the authenticated bot has access to.
func discordDelThreadHandler(botSession *session.Session, m *gateway.ThreadDeleteEvent, botName string) {

}

// kick a user and log it to a channel if configured
// session, guild, user being kicked,
func kickDiscordUser(botSession *session.Session, guild discord.Guild, user discord.User, username, reason, authorname string) (err error) {
	if err = botSession.Kick(guild.ID, user.ID, api.AuditLogReason(reason)); err != nil {
		return
	}

	// TODO: Need to use new config for this
	// sendDiscordEmbed(getDiscordConfigString("embed.audit"), embed)

	Log.Info("User " + user.Username + " has been kicked from " + guild.Name + " for " + reason)

	return
}

// ban a user and log it to a channel if configured
func banDiscordUser(botSession *session.Session, guild discord.Guild, user discord.User, username, reason, authorname string, days int) (err error) {
	banData := api.BanData{
		DeleteDays:     option.NewUint(uint(days)),
		AuditLogReason: api.AuditLogReason(reason),
	}

	if err = botSession.Ban(guild.ID, user.ID, banData); err != nil {
		return
	}

	// TODO: Need to use new config for embed audit to log to a webhook

	Log.Info("User " + user.Username + " has been banned from " + guild.Name + " for " + reason)

	return
}

// clean up messages if configured to
func deleteDiscordMessages(botSession *session.Session, channel *discord.Channel, messages []discord.MessageID, reason string) (err error) {
	Log.Debugf("Removing messages from %s", channel.Name)

	if err = botSession.DeleteMessages(channel.ID, messages, api.AuditLogReason(reason)); err != nil {
		return
	}

	// TODO: Need to use new config for embed audit to log to a webhook

	Log.Debug("messages were deleted.")

	return
}

// send message handling
func sendDiscordMessage(botSession *session.Session, channel *discord.Channel, author discord.User, responseArray []string, botName string) (err error) {
	// if there is no response to send just return
	if len(responseArray) == 0 {
		return
	}

	prefix := getPrefix("discord", botName, channel.GuildID.String())

	response := strings.Join(responseArray, "\n")
	response = strings.Replace(response, "&user&", "<@"+author.ID.String()+">", -1)
	response = strings.Replace(response, "&prefix&", prefix, -1)
	response = strings.Replace(response, "&react&", "", -1)

	// if there is an error return the error
	if _, err = botSession.SendMessage(channel.ID, response); err != nil {
		return
	}

	return
}

// send a reaction to a message
func sendDiscordReaction(botSession *session.Session, channel *discord.Channel, message discord.Message, reactionArray []string) (err error) {
	// if there is no reaction to send just return
	if len(reactionArray) == 0 || len(reactionArray) == 1 && reactionArray[0] == "" {
		return
	}

	var hasReact bool

	for _, reaction := range reactionArray {
		if reaction != "" {
			hasReact = true
			break
		}
	}

	if !hasReact {
		return nil
	}

	for _, reaction := range reactionArray {
		Log.Debugf("sending \"%s\" as a reaction to message: %s", reaction, message.ID)
		// if there is an error sending a message return it
		if err = botSession.React(channel.ID, message.ID, discord.APIEmoji(reaction)); err != nil {
			return
		}
	}
	return
}

// send a message with an embed
func sendDiscordEmbed(botSession *session.Session, channel discord.Channel, embed *discord.Embed) error {
	// if there is an error sending the embed message
	if _, err := botSession.SendEmbeds(channel.ID, *embed); err != nil {
		Log.Fatal("Embed send error")
		return err
	}

	return nil
}

// service handling
// start all the bots
func startDiscordsBots() {
	Log.Infof("Starting discord server connections\n")
	// range over the bots available to start
	for _, bot := range discordGlobal.Bots {
		Log.Infof("Connecting to %s\n", bot.BotName)

		// spin up a channel to tell the bot to stop later
		// stopDiscord[bot.BotName] = make(chan string)

		// start the bot
		go startDiscordBotConnection(bot)
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
		stopDiscordServer <- bot.BotName
	}

	for range discordGlobal.Bots {
		botIn := <-discordServerStopped
		Log.Infof("%s", botIn)
	}

	Log.Infof("discord connections stopped")
	// return shutdown signal on channel
	servStopped <- "discord_stopped"
}

// start connections to discord
func startDiscordBotConnection(discordConfig discordBot) {
	Log.Debugf("starting connections for %s", discordConfig.BotName)
	// Initializing Discord connection

	// Create a new Discord session using the provided bot token.
	Log.Debugf("using token '%s' to auth", discordConfig.Config.Token)
	botSession := session.New("Bot " + discordConfig.Config.Token)

	// Add Gateway Intents
	botSession.AddIntents(gateway.IntentGuildMessages)
	botSession.AddIntents(gateway.IntentGuildEmojis)
	botSession.AddIntents(gateway.IntentGuildModeration)
	botSession.AddIntents(gateway.IntentDirectMessages)

	// Register ready as a callback for the ready event
	botSession.AddHandler(func(gate *gateway.ReadyEvent) {
		readyDiscord(botSession, discordConfig.Config.Game)
	})

	// Register messageCreate as a callback for the messageCreate events.
	for range discordConfig.Servers {
		botSession.AddHandler(func(gate *gateway.MessageCreateEvent) {
			discordMessageHandler(botSession, gate, discordConfig.BotName)
		})
	}

	for range discordConfig.Servers {
		botSession.AddHandler(func(gate *gateway.ThreadCreateEvent) {})
	}

	for range discordConfig.Servers {
		botSession.AddHandler(func(gate *gateway.ThreadDeleteEvent) {})
	}

	// Open the websocket and begin listening.
	if err := botSession.Open(context.Background()); err != nil {
		Log.Error("Failed to connect:", err)
	}

	Log.Debugf("Discord service connected for %s", discordConfig.BotName)

	//bot, err := dg.User("@me")
	botUser, err := botSession.Me()
	if err != nil {
		fmt.Println("error obtaining account details,", err)
		syscall.Exit(2)
	}

	// Permissions requested
	// Administration - Kick/Ban/Moderate Members
	// Text -
	//		Send/Manage Messages w/In Threads
	//		Create/Manage Threads
	//		Add Reactions
	//		User External Emojis
	//		Embed links

	Log.Debug("Invite the bot to your server with https://discordapp.com/oauth2/authorize?client_id=" + botUser.ID.String() + "&scope=bot&permissions=1495185845318")

	discordLoad <- ""

	<-stopDiscordServer

	Log.Debugf("stop recieved on %s", discordConfig.BotName)

	if err := botSession.Close(); err != nil {
		Log.Error(err)
	}

	Log.Debugf("%s sent close", discordConfig.BotName)
	// return the shutdown signal
	discordServerStopped <- fmt.Sprintf("Closed connection for %s", discordConfig.BotName)
}
