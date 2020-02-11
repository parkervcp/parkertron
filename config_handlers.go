package main

func getBlacklist(inService, botName, inServer, inChannel string) (blacklist []string) {
	perms := []permission{}

	switch inService {
	case "discord":
		for _, bot := range discordGlobal.Bots {
			if bot.BotName == botName {
				for _, server := range bot.Servers {
					if inServer == server.ServerID {
						perms = server.Permissions
					}
				}
			}
		}
	case "irc":
		for _, group := range getChannelGroups(inService, botName, inServer, inChannel) {
			for _, channel := range group.ChannelIDs {
				if channel == inChannel {
					perms = group.Permissions
				}
			}
		}
	default:
	}

	// load users that are in blacklisted groups
	for _, perm := range perms {
		if perm.Blacklisted {
			for _, user := range perm.Users {
				blacklist = append(blacklist, user)
			}
		}
	}

	return
}

func getChannels(inService, botName, inServer string) (channels []string) {
	switch inService {
	case "discord":
		for _, bot := range discordGlobal.Bots {
			if bot.BotName == botName {
				for _, server := range bot.Servers {
					if inServer == server.ServerID {
						for _, group := range server.ChanGroups {
							for _, channel := range group.ChannelIDs {
								channels = append(channels, channel)
							}
						}
					}
				}
			}
		}
	case "irc":
		for _, bot := range ircGlobal.Bots {
			if bot.BotName == botName {
				for _, group := range bot.ChanGroups {
					for _, channel := range group.ChannelIDs {
						channels = append(channels, channel)
					}
				}
			}
		}
	default:
	}

	return
}

func getChannelGroups(inService, botName, inServer, inChannel string) (chanGroups []channelGroup) {
	switch inService {
	case "discord":
		for _, bot := range discordGlobal.Bots {
			if bot.BotName == botName {
				for _, server := range bot.Servers {
					if inServer == server.ServerID {
						chanGroups = server.ChanGroups
					}
				}
			}
		}
	case "irc":
		for _, bot := range ircGlobal.Bots {
			if bot.BotName == botName {
				chanGroups = bot.ChanGroups
			}
		}
	default:
	}

	return
}

func getCommands(inService, botName, inServer, inChannel string) (commands []command) {
	// prep stuff for passing to the parser
	for _, group := range getChannelGroups(inService, botName, inServer, inChannel) {
		for _, channel := range group.ChannelIDs {
			if inChannel == channel {
				for _, command := range group.Commands {
					commands = append(commands, command)
				}
			}
		}
	}

	return
}

func getKeywords(inService, botName, inServer, inChannel string) (keywords []keyword) {
	// prep stuff for passing to the parser
	for _, group := range getChannelGroups(inService, botName, inServer, inChannel) {
		for _, channel := range group.ChannelIDs {
			if inChannel == channel {
				for _, keyword := range group.Keywords {
					keywords = append(keywords, keyword)
				}
			}
		}
	}

	return
}

func getMentions(inService, botName, inServer, inChannel string) (ping, mention responseArray) {
	switch inService {
	case "discord":
		for _, bot := range discordGlobal.Bots {
			if bot.BotName == botName {
				for _, server := range bot.Servers {
					if inServer == server.ServerID {
						if inChannel == "DirectMessage" {
							mention = bot.Config.DMResp
						} else {
							for _, group := range server.ChanGroups {
								for _, channel := range group.ChannelIDs {
									if inChannel == channel {
										Log.Debugf("bot was mentioned on channel %s", channel)
										Log.Debugf("ping resp %s", group.Mentions.Ping)
										Log.Debugf("mention resp %s", group.Mentions.Mention)
										ping = group.Mentions.Ping
										mention = group.Mentions.Mention
										return
									}
								}
							}
						}
					}
				}
			}
		}
	case "irc":
		for _, bot := range ircGlobal.Bots {
			if bot.BotName == botName {
				if inChannel == bot.Config.Server.Nickname {
					mention = bot.Config.DMResp
				} else {
					for _, group := range bot.ChanGroups {
						for _, channel := range group.ChannelIDs {
							if inChannel == channel {
								ping = group.Mentions.Ping
								mention = group.Mentions.Mention
								return
							}
						}
					}
				}
			}
		}
	default:
	}

	return
}

func getParsing(inService, botName, inServer, inChannel string) (parseConf parsing) {
	// prep stuff for passing to the parser
	for _, group := range getChannelGroups(inService, botName, inServer, inChannel) {
		for _, channel := range group.ChannelIDs {
			if inChannel == channel {
				parseConf = group.Parsing
			}
		}
	}

	return
}

func getBotParseConfig() (maxLogs int, response, reaction []string, allowIP bool) {
	return botConfig.Parsing.Max, botConfig.Parsing.Response, botConfig.Parsing.Reaction, botConfig.Parsing.AllowIP
}

func getPrefix(inService, botName, inServer, inChannel string) (prefix string) {
	switch inService {
	case "discord":
		for _, bot := range discordGlobal.Bots {
			if bot.BotName == botName {
				for _, server := range bot.Servers {
					if inServer == server.ServerID {
						prefix = server.Config.Prefix
					}
				}
			}
		}
	case "irc":
		for _, bot := range ircGlobal.Bots {
			if bot.BotName == botName {
				prefix = bot.Config.Prefix
			}
		}
	default:
	}

	return
}
