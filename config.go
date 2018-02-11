package main

import (
	"github.com/spf13/viper"
)

var ()

type viperConfig struct {
}

func setConfig() {

	discord := viper.New()

	discord.SetConfigName("discord")
	discord.AddConfigPath("configs/")
	discord.WatchConfig()

	if err := discord.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}
}

func botConfig() {
	bot := viper.New()
	discord := viper.New()

	bot.SetConfigName("bot")
	bot.AddConfigPath("configs/")
	bot.WatchConfig()

	if err := bot.ReadInConfig(); err != nil {
		writeLog("fatal", "Could not load configuration.", err)
		return
	}
}
