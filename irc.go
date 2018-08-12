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

//ircMessageHandler the IRC listener that manages inbound messaging
func ircMessageHandler() {
	message, err := c.ReadMessage()
	if err != nil {
		fatal("cannot read message: ", err)
		return
	}

	debug("irc inbound " + message.String())

	if message.Command == "PING" {
		c.Send("PONG " + message.Trailing)
		debug("PONG Sent")
		return
	}

	if message.Command == "NOTICE" {
		if strings.Contains(strings.ToLower(message.Trailing), "this nickname is registered") {
			c.Send("%s IDENTIFY %s %s", message.Nick(), getIRCConfigString("nick"), getIRCConfigString("password"))
		}
		return
	}

	if message.Command == "PRIVMSG" {
		input := message.Trailing

		if message.Nick() == getIRCConfigString("nick") || strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
			if message.Nick() == getIRCConfigString("nick") {
				debug("User is the bot and being ignored.")
				return
			}
			if strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
				debug("User is blacklisted")
				return
			}
		}

		if message.Params[0] == getIRCConfigString("nick") {
			sendIRCMessage(message.Nick(), "Thank you for messaging me, but I only offer support in the main chat.")
			return
		}

		//
		// Message Handling
		//
		if input != "" {
			debug("Message Content: " + input)

			if strings.HasPrefix(input, getIRCConfigString("prefix")) == false {
				debug("sending to \"" + message.Params[0])
				parseKeyword("irc", message.Params[0], input)
				return
			} else if strings.HasPrefix(input, getIRCConfigString("prefix")) == true {
				input := strings.TrimPrefix(input, getIRCConfigString("prefix"))
				debug("sending to \"" + message.Params[0])
				parseCommand("irc", message.Params[0], message.Nick(), input)
				return
			}
			return
		}
		debug(message.Raw)
	}
}

//sendIRCMessage function to send messages separate of the listener
func sendIRCMessage(ChannelID string, response string) {
	response = strings.Replace(response, "&prefix&", getIRCConfigString("prefix"), -1)
	multiresp := strings.Split(response, "\n")

	debug("IRC Message Sent:")

	for x := range multiresp {
		debug("line sent: " + multiresp[x])
		c.Send("PRIVMSG " + ChannelID + " :" + multiresp[x])
	}
}

func startIRCConnection() {
	// This is the address of the irc server and the port combined to make it easier to input later
	address = getIRCConfigString("server.address") + ":" + getIRCConfigString("server.port")

	debug("Address should be " + getIRCConfigString("server.address") + ":" + getIRCConfigString("server.port"))

	debug("Connecting on " + address)

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

	debug("irc service started\n")

	ServStat <- "irc_online"

	for {
		ircMessageHandler()
	}
}
