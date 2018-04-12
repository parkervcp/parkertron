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

	//Setting Bot config settings
	Bot.SetConfigName("bot")
	Bot.AddConfigPath("configs/")
	Bot.WatchConfig()

	Bot.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Bot config changed", nil)
	})

	if err := Bot.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load Bot configuration.", err)
		return
	}

	//Setting Discord config settings
	Discord.SetConfigName("discord")
	Discord.AddConfigPath("configs/")
	Discord.WatchConfig()

	Discord.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Discord config changed", nil)
	})

	if err := Discord.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load Discord configuration.", err)
		return
	}

	//Setting IRC config settings
	IRC.SetConfigName("irc")
	IRC.AddConfigPath("configs/")
	IRC.WatchConfig()

	IRC.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "IRC config changed", nil)
	})

	if err := IRC.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load irc configuration.", err)
		return
	}

	//Setting Command config settings
	Command.SetConfigName("command")
	Command.AddConfigPath("configs/")
	Command.WatchConfig()

	Command.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Command config changed", nil)
	})

	if err := Command.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load Command configuration.", err)
		return
	}

	//Setting Keyword config settings
	Keyword.SetConfigName("keywords")
	Keyword.AddConfigPath("configs/")
	Keyword.WatchConfig()

	Keyword.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Keyword config changed", nil)
	})

	if err := Keyword.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load Keyword configuration.", err)
		return
	}

	//Setting Keyword config settings
	Parsing.SetConfigName("parsing")
	Parsing.AddConfigPath("configs/")
	Parsing.WatchConfig()

	Parsing.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Parsing config changed", nil)
	})

	if err := Parsing.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load Parsing configuration.", err)
		return
	}
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

func getDiscordGroupMembers(req string) string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.group."+req), " "))
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
