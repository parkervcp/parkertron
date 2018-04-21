package main

import (
	"log"
	"strings"
	"time"

	"github.com/husio/irc"
)

var (
	address = getIRCConfigString("server.address") + ":" + getIRCConfigString("server.port")

	c, err = irc.Connect(address)
)

func ircMessageHandler() {
	message, err := c.ReadMessage()
	if err != nil {
		writeLog("fatal", "cannot read message: ", err)
		return
	}

	writeLog("debug", "irc inbound "+message.String(), nil)

	if message.Command == "PING" {
		c.Send("PONG " + message.Trailing)
		writeLog("debug", "PONG Sent", nil)
		return
	}

	if message.Command == "NOTICE" {
		if strings.Contains(strings.ToLower(message.Trailing), "this nickname is registered") {
			c.Send("%s IDENTIFY %s %s", message.Nick(), getIRCConfigString("nick"), getIRCConfigString("password"))
		}
		return
	}

	response = ""
	multiresp := []string{}

	if message.Command == "PRIVMSG" {
		input := message.Trailing

		if message.Nick() == getIRCConfigString("nick") || strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
			if message.Nick() == getIRCConfigString("nick") {
				writeLog("debug", "User is the bot and being ignored.", nil)
			}
			if strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
				writeLog("debug", "User is blacklisted", nil)
			}
			return
		}

		if message.Params[0] == getIRCConfigString("nick") {
			response = "Thank you for messaging me, but I only offer support in the main chat."
			c.Send("PRIVMSG %s :%s \"%s\"", message.Nick(), message.Nick(), response)
			return
		}

		//
		// Message Handling
		//
		if input != "" {
			writeLog("debug", "Message Content: "+input+"\n", nil)

			if strings.HasPrefix(input, getIRCConfigString("prefix")) == false {
				response = parseKeyword(input)

			} else if strings.HasPrefix(input, getIRCConfigString("prefix")) == true {
				trimmed := strings.TrimPrefix(input, getIRCConfigString("prefix"))
				response = parseCommand(trimmed)

				if response == "" {
					return
				}

			} else {
				response = "That's not a recognized command."
			}

			if response == "" {
				return
			}

			writeLog("debug", "Message Sent: \n"+response+"\n", nil)

			multiresp = strings.Split(response, "\n")

		}

		writeLog("debug", message.Raw, nil)

		for x := range multiresp {
			writeLog("debug", multiresp[x], nil)
			c.Send("PRIVMSG %s :%s %s", message.Params[0], message.Nick(), multiresp[x])
		}
	}
}

func ircAuth() {
	message, err := c.ReadMessage()
	if err != nil {
		log.Fatalf("cannot read message: %s", err)
		return
	}

	// check if registered
	c.Send("PRIVMSG NickServ info")
	if message.Command == "PRIVMSG" {
		if message.Params[0] == "NickServ" {
			// If unregistered start the registration process.
			if strings.Contains(message.Trailing, getIRCConfigString("nick")+" is not registered") == true {
				writeLog("debug", "Registereing IRC bot", nil)
				c.Send("PRIVMSG NickServ register %s %s", getIRCConfigString("password"), getIRCConfigString("email"))
			}
			// If Registered log in as the user
			if strings.Contains(message.Trailing, "You are now identified for "+getIRCConfigString("nick")) {
				return
			}
		}
	}
}

func startIRCConnection() {
	// This is the address of the irc server and the port combined to make it easier to input later
	address = getIRCConfigString("server.address") + ":" + getIRCConfigString("server.port")

	writeLog("debug", "Address should be "+getIRCConfigString("server.address")+":"+getIRCConfigString("server.port"), nil)

	writeLog("debug", "Connecting on "+address, nil)

	c, err = irc.Connect(address)

	// Connect to the server
	if err != nil {
		log.Fatalf("cannot connect to %q: %s", address, err)
	}

	// send user info
	c.Send("USER %s %s * :"+getIRCConfigString("real"), getIRCConfigString("ident"), address)
	c.Send("NICK %s", getIRCConfigString("nick"))

	time.Sleep(time.Millisecond * 50)

	for _, name := range getIRCChannels() {
		if !strings.HasPrefix(name, "#") {
			name = "#" + name
		}
		c.Send("JOIN %s", name)
	}

	writeLog("info", "irc service started\n", nil)

	ServStat <- "irc_online"

	for {
		ircMessageHandler()
	}
}
