package main

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	//Bot base config
	Bot = viper.New()
	//Discord Config for the bot
	Discord = viper.New()

	//Group config
	Group = viper.New()
)

func setConfig() {

	Bot.SetConfigName("bot")
	Bot.AddConfigPath("configs/")
	Bot.WatchConfig()

	Bot.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Bot config changed", nil)
	})

	if err := Bot.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}

	Discord.SetConfigName("discord")
	Discord.AddConfigPath("configs/")
	Discord.WatchConfig()

	Discord.OnConfigChange(func(e fsnotify.Event) {
		writeLog("info", "Discord config changed", nil)
	})

	if err := Discord.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}
}

//Bot Get funcs

func getBotServices() string {
	return strings.Join(Bot.GetStringSlice("services"), " ")
}

func getBotConfigString(req string) string {
	return Bot.GetString(req)
}

func getBotConfigInt(req string) int {
	return Bot.GetInt(req)
}

func getBotConfigBool(req string) bool {
	return Bot.GetBool(req)
}

//Discord get func

func getDiscordConfigString(req string) string {
	return Discord.GetString(req)
}

func getDiscordConfigInt(req string) int {
	return Discord.GetInt(req)
}

func getDiscordConfigBool(req string) bool {
	return Discord.GetBool(req)
}

func getDiscordChannels() string {
	return strings.Join(Discord.GetStringSlice("discord.channels"), " ")
}

func getDiscordGroupMembers(req) string {
	return strings.Join(Discord.GetStringSlice("discord.group."+req), " ")
}
