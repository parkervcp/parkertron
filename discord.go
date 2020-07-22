package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"mvdan.cc/xurls/v2"
)

var (
	stopDiscordServer    = make(chan string)
	discordServerStopped = make(chan string)

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
// message is created on any channel that the authenticated bot has access to.
func discordMessageHandler(dg *discordgo.Session, m *discordgo.MessageCreate, botName string) {
	Log.Debugf("bot is %s", botName)
	Log.Debugf("message '%s'", m.Content)

	// data to send to discord
	var response []string
	var reaction []string

	bot, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Ignore all messages created by bots (stops the bot uprising)
	if m.Author.Bot {
		Log.Debug("User is a bot and being ignored.")
		return
	}

	// get channel information
	channel, err := dg.State.Channel(m.ChannelID)
	if err != nil {
		Log.Fatal("Channel error", err)
		return
	}

	// get all the configs
	// requires channel info we get from the channel info above
	prefix := getPrefix("discord", botName, channel.GuildID)
	channelCommands := getCommands("discord", botName, channel.GuildID, m.ChannelID)
	channelKeywords := getKeywords("discord", botName, channel.GuildID, m.ChannelID)
	channelParsing := getParsing("discord", botName, channel.GuildID, m.ChannelID)
	serverFilter := getFilter("discord", botName, channel.GuildID)

	Log.Debugf("prefix: %s", prefix)

	// bot level configs for log reading
	maxLogs, logResponse, logReaction, allowIP := getBotParseConfig()

	// if the channel is a DM
	if channel.Type == 1 {
		_, dmResp := getMentions("discord", botName, channel.GuildID, "DirectMessage")
		if err := sendDiscordMessage(dg, m.ChannelID, m.Author.ID, getPrefix("discord", botName, channel.GuildID), dmResp.Reaction); err != nil {
			Log.Error(err)
		}

		if err := sendDiscordReaction(dg, m.ChannelID, m.ID, dmResp.Reaction); err != nil {
			Log.Error(err)
		}

		return
	}

	// Log.Debugf("all channels %s", getChannels("discord", botName, channel.GuildID))

	//filter logic
	Log.Debug("filtering messages")
	if len(serverFilter) == 0 {
		Log.Debugf("no filtered terms found")
	} else {
		for _, filter := range serverFilter {
			if strings.Contains(m.Content, filter.Term) {
				Log.Infof("message was removed for containing %s", filter.Term)
				if err := deleteDiscordMessage(dg, m.ChannelID, m.ID, ""); err != nil {
					Log.Error(err)
				}

				if err := sendDiscordMessage(dg, m.ChannelID, m.Author.ID, prefix, filter.Reason); err != nil {
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
	if !contains(getChannels("discord", botName, channel.GuildID), m.ChannelID) {
		Log.Debugf("channel not found")
		return
	}

	Log.Debugf("checking blacklist")

	// drop messages from blacklisted users
	for _, user := range getBlacklist("discord", botName, channel.GuildID, m.ChannelID) {
		if user == m.Author.ID {
			Log.Debugf("user %s is blacklisted username is %s", m.Author.ID, m.Author.Username)
			return
		}
	}

	Log.Debugf("checking attachments")

	// for all attachment urls
	var attachmentURLs []string
	for _, url := range m.Attachments {
		attachmentURLs = append(attachmentURLs, url.ProxyURL)
	}

	// this was for debugging/testing only
	// for _, url := range attachmentURLs {
	// 	parseURL(url, channelParsing)
	// }

	Log.Debugf("all attachments %s", attachmentURLs)
	Log.Debugf("all ignores %+v", channelParsing.Paste.Ignore)

	Log.Debugf("checking for any urls in the message")
	var allURLS []string
	for _, url := range xurls.Relaxed().FindAllString(m.Content, -1) {
		Log.Debugf("checking on %s", url)
		// if the url is an ip filter it out
		if match, err := regexp.Match("^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])", []byte(url)); err != nil {
			Log.Error(err)
		} else if match && !allowIP {
			Log.Debugf("adding %s to the list", url)
			continue
		}

		Log.Debugf("looking for ignored domains")
		if len(channelParsing.Paste.Ignore) == 0 {
			Log.Debugf("appending %s to allURLS", url)
			allURLS = append(allURLS, url)
			Log.Debugf("no ignored domain found")
			continue
		} else {
			var ignored bool
			for _, ignoreURL := range channelParsing.Paste.Ignore {
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

	// this was for debugging/testing only
	// for _, url := range allURLS {
	// 	parseURL(url, channelParsing)
	// }
	// Log.Debug(allURLS)

	// add all urls together
	Log.Debug("adding attachment URLS to allURLS")
	for i := 0; i < len(attachmentURLs); i++ {
		allURLS = append(allURLS, attachmentURLs[i])
	}

	// Log.Debug(allURLS)
	Log.Debugf("checking mentions")
	if len(m.Mentions) != 0 {
		ping, mention := getMentions("discord", botName, channel.GuildID, m.ChannelID)
		if m.Mentions[0].ID == bot.ID && strings.Replace(m.Content, "<@!"+dg.State.User.ID+">", "", -1) == "" {
			Log.Debugf("bot was pinged")
			response = ping.Response
			reaction = ping.Reaction
		} else {
			for _, mentioned := range m.Mentions {
				if mentioned.ID == bot.ID {
					Log.Debugf("bot was mentioned")
					response = mention.Response
					reaction = mention.Reaction
				}
			}
		}
	} else {
		Log.Debugf("no mentions found")
		if strings.HasPrefix(m.Content, prefix) {
			// command
			response, reaction = parseCommand(strings.TrimPrefix(m.Content, prefix), botName, channelCommands)
			// if the flag for clearing commands is set and there is a response
			if getCommandClear("discord", botName, channel.GuildID) && len(response) > 0 {
				Log.Debugf("removing command message %s", m.ID)
				if err := deleteDiscordMessage(dg, m.ChannelID, m.ID, ""); err != nil {
					Log.Error(err)
				}
			} else {

			}
		} else {
			// keyword
			response, reaction = parseKeyword(m.Content, botName, channelKeywords, channelParsing)
		}
	}

	if len(channelParsing.Image.FileTypes) == 0 && len(channelParsing.Paste.Sites) == 0 {
		Log.Debugf("no parsing configs found")
	} else {
		Log.Debugf("allURLS: %s", allURLS)
		Log.Debugf("allURLS count: %d", len(allURLS))

		// if we have too many logs ignore it.
		if len(allURLS) == 0 {
			Log.Debugf("no URLs to read")
		} else if len(allURLS) > maxLogs {
			Log.Debug("too many logs or screenshots to try and read.")
			if err := sendDiscordMessage(dg, m.ChannelID, m.Author.ID, prefix, logResponse); err != nil {
				Log.Error(err)
			}
			if err := sendDiscordReaction(dg, m.ChannelID, m.ID, logReaction); err != nil {
				Log.Error(err)
			}
			return
		} else {
			Log.Debugf("reading logs")
			if err := sendDiscordReaction(dg, m.ChannelID, m.ID, []string{"👀"}); err != nil {
				Log.Error(err)
			}

			// get parsed content for each url/attachment
			Log.Debugf("reading all attachments and logs")
			allParsed := make(map[string]string)
			for _, url := range allURLS {
				allParsed[url] = parseURL(url, channelParsing)
			}

			//parse logs and append to current response.
			for _, url := range allURLS {
				Log.Debugf("passing %s to keyword parser", url)
				urlResponse, _ := parseKeyword(allParsed[url], botName, channelKeywords, channelParsing)
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
	Log.Debugf("sending response %s to %s", response, m.ChannelID)
	if err := sendDiscordMessage(dg, m.ChannelID, m.Author.ID, prefix, response); err != nil {
		Log.Error(err)
	}

	// send reaction to channel
	Log.Debugf("sending reaction %s", reaction)
	if err := sendDiscordReaction(dg, m.ChannelID, m.ID, reaction); err != nil {
		Log.Error(err)
	}
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
	response = strings.Replace(response, "&user&", "<@"+authorID+">", -1)
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
	Log.Infof("Starting discord server connections\n")
	// range over the bots available to start
	for _, bot := range discordGlobal.Bots {
		Log.Infof("Connecting to %s\n", bot.BotName)

		// spin up a channel to tell the bot to shutdown later
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
	for range discordConfig.Servers {
		dg.AddHandler(func(dg *discordgo.Session, event *discordgo.MessageCreate) {
			discordMessageHandler(dg, event, discordConfig.BotName)
		})
	}

	Log.Debugf("Discord service connected for %s", discordConfig.BotName)

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

	<-stopDiscordServer

	Log.Debugf("stop recieved on %s", discordConfig.BotName)
	// properly send a shutdown to the discord server so the bot goes offline.
	if err := dg.Close(); err != nil {
		Log.Error(err)
	}

	Log.Debugf("%s sent close", discordConfig.BotName)
	// return the shutdown signal
	discordServerStopped <- fmt.Sprintf("Closed connection for %s", discordConfig.BotName)
}
