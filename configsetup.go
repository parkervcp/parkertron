package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tcnksm/go-input"
)

// configFilecheck to check if the files exist before starting up.
func configFilecheck() bool {
	if _, err := os.Stat("configs/"); err != nil {
		// create configs folder
		if os.IsNotExist(err) {
			os.Mkdir("configs/", 0755)
		}
	} else {
		info("Config folder exists")
		if checkConfigExists("bot.yml") == false {
			time.Sleep(5 * time.Second)
			setupBotConfig()
		}
		for _, cr := range getBotServices() {
			if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) == true {
				if strings.Contains(cr, "discord") == true {
					if checkConfigExists("discord.yml") == false {
						setupDiscordConfig()
					}
				}
				if strings.Contains(cr, "irc") == true {
					if checkConfigExists("irc.yml") == false {
						setupIRCConfig()
					}
				}
			}
		}

		if checkConfigExists("commands.yml") == false {
			setupCommandsConfig()
		}
		if checkConfigExists("keywords.yml") == false {
			setupKeywordsConfig()
		}
		return true
	}

	return false
}

func checkConfigExists(file string) bool {
	if _, err := os.Stat("configs/" + file); err == nil {
		return true
	}
	info("Need to generate the " + file + ". Doing that now.")
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
			} else if in == "N" || in == "n" || in == "No" || in == "no" {
				answer = false
			} else {
				return fmt.Errorf("Need to have a Y/n answer")
			}
			return nil
		},
	})
	if err != nil {
		fatal("", err)
	}

	return answer
}

func askStringQuestion(question string) string {
	ui := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}

	answer, err := ui.Ask(question, &input.Options{
		// Read the default val from env var
		Loop: true,
	})
	if err != nil {
		fatal("", err)
	}

	return answer
}

func setupBotConfig() {
	var services []string

	if askBoolQuestion("Do you plan on supporting discord? [Y/n]") == true {
		services = append(services, "discord")
		info("Discord enabled")
		setupDiscordConfig()
	}
	if askBoolQuestion("Do you plan on supporting irc? [Y/n]") == true {
		services = append(services, "irc")
		info("IRC enabled")
		setupIRCConfig()
	}

	Bot.Set("bot.services", services)

	if askBoolQuestion("Do you want to use the default log location of 'logs/'? [Y/n]") == false {
		Bot.Set("bot.log.location", askStringQuestion("Where you like logs to be saved?"))
	} else {
		Bot.Set("bot.log.location", "logs/")
	}

	if askBoolQuestion("Do you want to use the default log level of info? [Y/n]") == false {
		level := askStringQuestion("what log level would you like to use? [info/debug]")
		if level == "info" || level == "debug" {
			Bot.Set("bot.log.level", level)
		} else {
			info("Invalid log level set with " + level + ". Defaulting to info")
			Bot.Set("bot.log.level", "info")
		}

	} else {
		Bot.Set("bot.log.level", "info")
	}

	Bot.WriteConfigAs("configs/bot.yml")
}

func setupDiscordConfig() {
	// get discord bot token
	Discord.Set("discord.token", askStringQuestion("What is your discord token?"))

	// set bot prefix
	if askBoolQuestion("Would you like to use the default prefix? '.' [Y/n]") == false {
		Discord.Set("discord.prefix", askStringQuestion("What is the prefix you'd like to use?"))
	} else {
		Discord.Set("discord.prefix", ".")
	}

	// set bot owner
	if askBoolQuestion("Would you like to set a owner? (defaults to server owner) [Y/n]") == true {
		Discord.Set("discord.owner", askStringQuestion("What is the server owners discord ID?"))
	} else {
		info("defaulting to server owner")
	}

	var listening []string

	// set channel filter up
	if askBoolQuestion("Do you want the bot to listen on specific channels? [Y/n]") == true {
		if askBoolQuestion("A channel to listen on it required. Would you like to set one now? [Y/n]") == true {
			listening = append(listening, askStringQuestion("What is the ID of the channel you want to listen on?"))
			Discord.Set("bot.channels.filter", true)
			Discord.Set("bot.channels.listening", listening)
			info("The channel is not verified will only work if correct.")
		} else {
			Discord.Set("bot.channels.filter", false)
			Discord.Set("bot.channels.listening", listening)
		}
	}

	/*
		TODO:
			Discord.Set("irc.channels.groups.admin", admin)
			Discord.Set("irc.channels.groups.mods", mods)
			Discord.Set("irc.channels.groups.blacklist", blacklist)
	*/

	Discord.WriteConfigAs("configs/discord.yml")
}

func setupIRCConfig() {
	// get irc server settings
	IRC.Set("irc.server.address", askStringQuestion("What is the server address"))
	IRC.Set("irc.server.port", askStringQuestion("What is the server port?"))
	IRC.Set("irc.server.ssl", askBoolQuestion("Does this connection require ssl [Y/n]"))

	// get irc user settings
	IRC.Set("irc.ident", askStringQuestion("What is the IRC username you are using?"))
	IRC.Set("irc.email", askStringQuestion("What is the email associated with the account?"))
	IRC.Set("irc.password", askStringQuestion("What is the password the bot is using?"))
	IRC.Set("irc.nick", askStringQuestion("What is the nick that the bot should use?"))
	IRC.Set("irc.real", askStringQuestion("What is the \"Real Name\" the bot should use?"))

	// get irc text parsing settings
	// set bot prefix
	if askBoolQuestion("Would you like to use the default prefix? '.' [Y/n]") == false {
		IRC.Set("irc.prefix", askStringQuestion("What is the prefix you'd like to use?"))
	} else {
		IRC.Set("irc.prefix", ".")
	}
	// get irc channels

	var listening []string

	if askBoolQuestion("Do you want to add channels to join now? (you can pm the bot) [Y/n]") == true {
		listening = append(listening, askStringQuestion("What channel is the bot supposed to join? (Without the # in the name)"))
		for askBoolQuestion("Do you want to add more channels to join now? [Y/n]") == true {
			listening = append(listening, askStringQuestion("What channel is the bot supposed to join? (Without the # in the name)"))
		}
	}

	IRC.Set("irc.channels.listening", listening)

	/*
		TODO:
			IRC.Set("irc.channels.groups.admin", admin)
			IRC.Set("irc.channels.groups.mods", mods)
			IRC.Set("irc.channels.groups.blacklist", blacklist)
	*/
	IRC.WriteConfigAs("configs/irc.yml")
}

func setupCommandsConfig() {
	var command string
	commandmap := make(map[string]interface{})
	exit := false

	if askBoolQuestion("Do you want to set up custom commands now? [Y/n]: ") == false {
		info("Writing default commands config to file")
	} else {
		for exit == false {
			command = askStringQuestion("What is the command you want to add? (It can have spaces in it ex: 'help command') (leave blank to stop adding commands): ")
			if command == "" {
				exit = true
			} else {
				commandmap[command] = setupStringMap()
			}
		}
	}

	Command.Set("command", commandmap)

	Command.WriteConfigAs("configs/commands.yml")
}

func setupKeywordsConfig() {
	var keyword string
	keywordmap := make(map[string]interface{})
	exit := false

	if askBoolQuestion("Do you want to set up custom keywords now? [Y/n]: ") == false {
		info("Writing default keywords config to file")
	} else {
		for exit == false {
			keyword = askStringQuestion("What is the keyword you want to add? (It can have spaces in it ex: 'i need help') (leave blank to stop adding commands): ")
			if keyword == "" {
				exit = true
			} else {
				keywordmap[keyword] = setupStringMap()
			}
		}
	}

	Keyword.Set("keyword", keywordmap)

	Keyword.WriteConfigAs("configs/keywords.yml")
}

func setupGroup() {
	/*

		TODO:
			This block is for future work. Namely permissions and other things.

			// get irc groups

			var group []string

			if askBoolQuestion("Do you want to add users to admin group now?? [Y/n]") == true {
				admin = append(admin, askStringQuestion(""))
				for askBoolQuestion("Do you want to add more groups to join now? [Y/n]") == true {
					admin = append(admin, askStringQuestion(""))
				}
			}
	*/
}

func setupStringMap() []string {
	var array []string
	var line string
	exit := false

	fmt.Println("Multi-line responses are supported, so we will keed adding lines until you specify to stop (blank response)")

	for exit == false {
		line = askStringQuestion("What do you want this line to say? (leave blank to exit): ")
		if line == "" {
			exit = true
		} else {
			array = append(array, line)
		}
	}

	return array
}
