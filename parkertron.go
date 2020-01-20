package main

import (
	"flag"
	"strings"

	Log "github.com/sirupsen/logrus"
)

var (
	//response is the bot response on the channel
	response string

	//ServStat is the Service Status channel
	ServStat = make(chan string)

	title = `
	                    __                  __                        
	___________ _______|  | __ ____________/  |________  ____   ____  
	\____ \__  \\_  __ \  |/ // __ \_  __ \   __\_  __ \/  _ \ /    \ 
	|  |_> > __ \|  | \/    <\  ___/|  | \/|  |  |  | \(  <_> )   |  \
	|   __(_____/|__|  |__|_ \\____)|__|   |__|  |__|   \____/|___|__/
	|__| v.0.1.0`
)

func init() {
	verbose := flag.String("v", "info", "set the console verbosity of the bot")
	flag.Parse()

	Log.Info(title)

	setupConfig()

	if *verbose == "debug" {
		setLogLevel("debug")
	} else if *verbose == "superdebug" {
		setLogLevel("debug")
		setBotConfigString("log.level", "superdebug")
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
		if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) {
			services = services + cr
		}
	}

	Log.Debug("services loaded are " + services)
}

func main() {
	for _, cr := range getBotServices() {
		if strings.Contains(strings.TrimPrefix(cr, "bot.services."), cr) {
			if strings.Contains(cr, "discord") {
				Log.Info("Starting Discord connector\n")
				go startDiscordConnection()
			}

			if strings.Contains(cr, "irc") {
				Log.Info("Starting IRC connector\n")
				go startIRCConnection()
			}
		}
	}

	for range getBotServices() {
		<-ServStat
	}

	Log.Debug("Commands being loaded are: " + getCommandsString())
	Log.Debug("Keywords being loaded are: " + getKeywordsString())

	Log.Info("Bot is now running. Press CTRL-C to exit.\n")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}
