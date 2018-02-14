package main

import (
	"flag"
	"strings"
)

var (
	//response is the bot response on the channel
	response string
)

func init() {
	verbose := flag.String("v", "info", "set the console verbosity of the bot")
	flag.Parse()
	if *verbose == "debug" {
		setLogLevel("debug")
	} else {
		setLogLevel("info")
	}

	setupConfig()

	setupLogger()

	writeLog("debug", "services loaded are "+getBotServices(), nil)
}

func main() {
	if strings.Contains(getBotServices(), "discord") == true {
		writeLog("info", "Starting discord connector\n", nil)
		go startDiscordConnection()
	}

	writeLog("debug", "Commands being loaded are: "+getCommandsString(), nil)
	writeLog("debug", "Keywords being loaded are: "+getKeywordsString(), nil)

	writeLog("info", "Bot is now running.  Press CTRL-C to exit.\n", nil)
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
}
