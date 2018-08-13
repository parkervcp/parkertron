package main

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	//Bot Config
	Bot = viper.New()

	//Discord Config
	Discord = viper.New()

	//IRC Config
	IRC = viper.New()

	//Command Config
	Command = viper.New()

	//Keyword Config
	Keyword = viper.New()

	//Parsing Config
	Parsing = viper.New()
)

func setupConfig() {

	if configFilecheck() == false {
		errmsg("There was an issue setting up the config", nil)
	}

	//Setting Bot config settings
	Bot.SetConfigName("bot")
	Bot.AddConfigPath("configs/")
	Bot.WatchConfig()

	Bot.OnConfigChange(func(e fsnotify.Event) {
		info("Bot config changed")
	})

	if err := Bot.ReadInConfig(); err != nil {
		fatal("Could not load Bot configuration.", err)
		return
	}

	for _, cr := range getBotServices() {
		if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) == true {
			if strings.Contains(cr, "discord") == true {
				//Setting Discord config settings
				Discord.SetConfigName("discord")
				Discord.AddConfigPath("configs/")
				Discord.WatchConfig()

				Discord.OnConfigChange(func(e fsnotify.Event) {
					info("Discord config changed")
				})

				if err := Discord.ReadInConfig(); err != nil {
					fatal("Could not load Discord configuration.", err)
					return
				}

				Discord.SetDefault("discord.command.remove", true)
			}

			if strings.Contains(cr, "irc") == true {
				//Setting IRC config settings
				IRC.SetConfigName("irc")
				IRC.AddConfigPath("configs/")
				IRC.WatchConfig()

				IRC.OnConfigChange(func(e fsnotify.Event) {
					info("IRC config changed")

				})
				if err := IRC.ReadInConfig(); err != nil {
					fatal("Could not load irc configuration.", err)
					return
				}
			}
		}
	}

	//Setting Command config settings
	Command.SetConfigName("commands")
	Command.AddConfigPath("configs/")
	Command.WatchConfig()

	Command.OnConfigChange(func(e fsnotify.Event) {
		info("Command config changed")
	})

	if err := Command.ReadInConfig(); err != nil {
		fatal("Could not load Command configuration.", err)
		return
	}

	//Setting Keyword config settings
	Keyword.SetConfigName("keywords")
	Keyword.AddConfigPath("configs/")
	Keyword.WatchConfig()

	Keyword.OnConfigChange(func(e fsnotify.Event) {
		info("Keyword config changed")
	})

	if err := Keyword.ReadInConfig(); err != nil {
		fatal("Could not load Keyword configuration.", err)
		return
	}

	//Setting website parsing config settings
	Parsing.SetConfigName("parsing")
	Parsing.AddConfigPath("configs/")
	Parsing.WatchConfig()

	Parsing.OnConfigChange(func(e fsnotify.Event) {
		info("Parsing config changed")
	})

	if err := Parsing.ReadInConfig(); err != nil {
		fatal("Could not load Parsing configuration.", err)
		return
	}

	info("Bot configs loaded")
}

//Bot Get funcs
func getBotServices() []string {
	return Bot.GetStringSlice("bot.services")
}

func getBotConfigBool(req string) bool {
	return Bot.GetBool("bot." + req)
}

func getBotConfigString(req string) string {
	return Bot.GetString("bot." + req)
}

func getBotConfigInt(req string) int {
	return Bot.GetInt("bot." + req)
}

func getBotConfigFloat(req string) float64 {
	return Bot.GetFloat64("bot." + req)
}

func setBotConfigString(req string, value string) {
	Bot.Set("bot."+req, value)
}

//Discord get funcs
func getDiscordConfigString(req string) string {
	return Discord.GetString("discord." + req)
}

func getDiscordConfigInt(req string) int {
	return Discord.GetInt("discord." + req)
}

func getDiscordConfigBool(req string) bool {
	return Discord.GetBool("discord." + req)
}

func getDiscordChannels() string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.channels.listening"), " "))
}

func getDiscordGroup(req string) []string {
	var groups []string
	for x := range Discord.GetStringMapString("discord.permissions.group") {
		groups = append(groups, x)
	}
	return groups
}

func getDiscordGroupMembers(req string) []string {
	var users []string
	for x := range Discord.GetStringMapString("discord.permissions.group." + req) {
		debug(x + " being added to the group")
		users = append(users, x)
	}
	return users
}

func getDiscordBlacklist() string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.permissions.group.blacklist"), " "))
}

func getDiscordKOMChannel(req string) bool {
	return Discord.IsSet("discord.kick_on_mention.channel." + req)
}

func getDiscordKOMID(req string) string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.kick_on_mention.channel."+req), " "))
}

func getDiscordKOMMessage(req string) string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.kick_on_mention.channel."+req+".message"), "\n"))
}

//IRC get funcs
func getIRCConfigString(req string) string {
	return IRC.GetString("irc." + req)
}

func getIRCConfigInt(req string) int {
	return IRC.GetInt("irc." + req)
}

func getIRCConfigBool(req string) bool {
	return IRC.GetBool("irc." + req)
}

func getIRCChannels() []string {
	return IRC.GetStringSlice("irc.channels.listening")
}

func getIRCGroupMembers(req string) string {
	return strings.ToLower(strings.Join(IRC.GetStringSlice("irc.group."+req), " "))
}

//Command get funcs
func getCommands() []string {
	return Command.AllKeys()
}

func getCommandsString() string {
	return strings.ToLower(strings.Replace(strings.Join(Command.AllKeys(), ", "), "command.", "", -1))
}

func getCommandResonse(req string) []string {
	return Command.GetStringSlice("command." + req)
}

func getCommandResponseString(req string) string {
	return strings.Join(Command.GetStringSlice("command."+req), "\n")
}

func getCommandStatus(req string) bool {
	for _, cr := range getCommands() {
		if strings.Contains(strings.TrimPrefix(cr, "command."), req) == true {
			return true
		}
	}
	return false
}

//Keyword get funcs
func getKeywords() []string {
	return Keyword.AllKeys()
}

func getKeywordsString() string {
	return strings.ToLower(strings.Replace(strings.Join(Keyword.AllKeys(), ", "), "keyword.", "", -1))
}

func getKeywordResponse(req string) []string {
	return Keyword.GetStringSlice("keyword." + req)
}

func getKeywordResponseString(req string) string {
	return strings.Join(Keyword.GetStringSlice("keyword."+req), "\n")
}

//Parsing get funcs
func getParsingPasteKeys() string {
	return strings.Replace(strings.Join(Parsing.AllKeys(), ", "), "parse.paste.", "", -1)
}

func getParsingPasteString(key string) string {
	return Parsing.GetString("parse.paste." + key)
}

func getParsingImageFiletypes() []string {
	return Parsing.GetStringSlice("parse.image.filetype")
}
