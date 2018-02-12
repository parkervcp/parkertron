package main

import (
	"strings"

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

	if err := Bot.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}

	Discord.SetConfigName("discord")
	Discord.AddConfigPath("configs/")
	Discord.WatchConfig()

	if err := Discord.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}
}

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

func getDiscordConfigString(req string) string {
	return Discord.GetString(req)
}

func getDiscordConfigInt(req string) int {
	return Discord.GetInt(req)
}
