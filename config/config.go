package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func getBotConf(bConfIn string) string {
	viper.SetConfigName("bot")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	// If no config is found, use the default(s)
	viper.SetDefault("botName", "parkertron")
	viper.SetDefault("actvCon", "console")
	viper.SetDefault("logsLoc", "logs/")

	bConfOut := viper.GetString(bConfIn)
	return bConfOut
}

func getDiscordConf(dConfIn string) string {
	viper.SetConfigName("discord")

	err := viper.ReadInConfig()

	if err != nil {
		fmt.Println("No configuration file loaded - using defaults")
	}

	// If no config is found, use the default(s)
	viper.SetDefault("cmdPrefix", ".")

	dConfOut := viper.GetString(dConfIn)

	return dConfOut
}

func saveConfigs() {
	// write out configs to files per service
}
