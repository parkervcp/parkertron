package main

import (
	"flag"
	"strings"
)

var (
	//response is the bot response on the channel
	response string

	//ServStat is the Service Status channel
	ServStat = make(chan string)
)

func init() {
	verbose := flag.String("v", "info", "set the console verbosity of the bot")
	flag.Parse()

	setupConfig()

	if *verbose == "debug" {
		setLogLevel("debug")
	} else {
		if getBotConfigString("log.level") == "" {
			setLogLevel("info")
		} else {
			setLogLevel(getBotConfigString("log.level"))
		}
	}

	setupLogger()

	services := ""

	for _, cr := range getBotServices() {
		if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) == true {
			services = services + cr
		}
	}

	writeLog("debug", "services loaded are "+services, nil)
}

func main() {
	for _, cr := range getBotServices() {
		if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) == true {
			if strings.Contains(cr, "discord") == true {
				writeLog("info", "Starting Discord connector\n", nil)
				go startDiscordConnection()
			}

			if strings.Contains(cr, "irc") == true {
				writeLog("info", "Starting IRC connector\n", nil)
				go startIRCConnection()
			}
		}
	}

	for range getBotServices() {
		<-ServStat
	}

	writeLog("debug", "Commands being loaded are: "+getCommandsString(), nil)
	writeLog("debug", "Keywords being loaded are: "+getKeywordsString(), nil)

	<-ServStat

	writeLog("info", "Bot is now running.  Press CTRL-C to exit.\n", nil)
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}
