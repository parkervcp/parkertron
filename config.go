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
)

func setConfig() {

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
}

//Bot Get funcs

func getBotServices() string {
	return strings.Join(Bot.GetStringSlice("services"), " ")
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

func getDiscordGroupMembers(req string) string {
	return strings.Join(Discord.GetStringSlice("discord.group."+req), " ")
}
