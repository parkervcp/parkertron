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

	//Command Config
	Command = viper.New()

	//Keyword Config
	Keyword = viper.New()
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

	//Setting Chat config settings
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
}

//Bot Get funcs
func getBotServices() string {
	return strings.ToLower(strings.Join(Bot.GetStringSlice("services"), " "))
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
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.channels"), " "))
}

func getDiscordGroupMembers(req string) string {
	return strings.ToLower(strings.Join(Discord.GetStringSlice("discord.group."+req), " "))
}

//Command get funcs
func getCommands() string {
	return strings.ToLower(strings.Join(Command.GetStringSlice("command"), ", "))
}

func getCommandResonse(req string) string {
	return strings.ToLower(strings.Join(Command.GetStringSlice("command."+req), "\n"))
}

//Chat get funcs
func getKeywords() string {
	return strings.ToLower(strings.Join(Keyword.GetStringSlice("keywords"), ", "))
}

func getKeywordResonse(req string) string {
	return strings.ToLower(strings.Join(Keyword.GetStringSlice("keywords."+req), "\n"))
}
