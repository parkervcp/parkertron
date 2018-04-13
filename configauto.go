package main

import (
	"fmt"
	"os"

	"github.com/tcnksm/go-input"
)

// configFilecheck to check if the files exist before starting up.
func configFilecheck() bool {
	if _, err := os.Stat("configs/"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("configs/", 0755)
		} else {
			// other error
		}
	} else {
		writeLog("info", "Config folder exists", nil)
		if checkConfigExists("bot.yml") == false {
			writeLog("info", "Need to generate config for the bot. Doing that now.", nil)
			setupBotConfig()
		}
		if checkConfigExists("discord.yml") == false {
			writeLog("info", "Need to generate config for discord. Doing that now.", nil)
		}
		return true
	}

	return false
}

func checkConfigExists(file string) bool {
	if _, err := os.Stat("configs/" + file); err == nil {
		return true
	}
	return false
}

func askBoolQuestion(question string) bool {

	var answer bool

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	_, err := ui.Ask(question, &input.Options{
		Required: true,
		Loop:     true,
		ValidateFunc: func(in string) error {
			if in == "Y" || in == "y" || in == "Yes" || in == "yes" {
				answer = true
				return nil
			} else if in == "N" || in == "n" || in == "No" || in == "no" {
				answer = false
				return nil
			} else {
				return fmt.Errorf("Need to have a Y/n answer")
			}
		},
	})
	if err != nil {
		writeLog("fatal", "", err)
	}

	return answer
}

func askStingQuestion(question string) string {

	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	answer, err := ui.Ask(question, &input.Options{
		Required: true,
		Loop:     true,
		// Validate input
		ValidateFunc: func(s string) error {
			if s != "Y" && s != "n" {
				writeLog("error", "", nil)
			}

			return nil
		},
	})
	if err != nil {
		writeLog("fatal", "", err)
	}

	return answer
}

func setupBotConfig() {
	var services []string

	discord := askBoolQuestion("Do you plan on supporting discord? [Y/n]")
	irc := askBoolQuestion("Do you plan on supporting irc? [Y/n]")

	switch {
	case discord == true:
		services = append(services, "discord")
	case irc == true:
		services = append(services, "irc")
	}

	servicestring := ""

	for _, cr := range services {
		servicestring = servicestring + cr + ", "
	}

	writeLog("debug", "services loaded are "+servicestring, nil)

	Bot.WriteConfigAs("configs/bot.yml")
}

func setupDiscordConfig() {

}

func setupIRCConfig() {

}

func setupCommandsConfig() {

}

func setupKeywordsConfig() {

}
