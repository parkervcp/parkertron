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

	// keep alive messaging
	if message.Command == "PING" {
		c.Send("PONG " + message.Trailing)
		debug("PONG Sent")
		return
	}

	// for authentication
	if message.Command == "NOTICE" {
		if strings.Contains(strings.ToLower(message.Trailing), "this nickname is registered") {
			c.Send("%s IDENTIFY %s %s", message.Nick(), getIRCConfigString("nick"), getIRCConfigString("password"))
		}
		return
	}

	// message handling
	if message.Command == "PRIVMSG" {
		dpack := DataPackage{}
		dpack.Service = "irc"
		dpack.Message = message.Trailing
		dpack.ChannelID = message.Params[0]
		dpack.AuthorID = message.Nick()
		dpack.BotID = getIRCConfigString("nick")

		superdebug("message.Params[0]: " + message.Params[0])
		superdebug("message.Nick(): " + message.Nick())
		superdebug("message.Trailing: " + message.Trailing)

		// if the user nickname matches bot or blacklisted.
		if message.Nick() == dpack.BotID || strings.Contains(getIRCBlacklist(), dpack.AuthorID) {
			if message.Nick() == dpack.BotID {
				debug("User is the bot and being ignored.")
				return
			}
			if strings.Contains(getIRCBlacklist(), dpack.AuthorID) {
				debug("User is blacklisted")
				return
			}
		}

		// if bot is DM'd
		if message.Params[0] == getIRCConfigString("nick") {
			debug("This was a DM")
			dpack.Response = getDiscordConfigString("direct.response")
			sendIRCMessage(message.Nick(), getIRCConfigString("direct.response"))
			return
		}

		//
		// Message Handling
		//
		if dpack.Message != "" {
			debug("Message Content: " + dpack.Message)

			if !strings.HasPrefix(dpack.Message, getIRCConfigString("prefix")) {
				debug("sending to \"" + message.Params[0])
				parseKeyword(dpack)
			} else {
				dpack.Message = strings.TrimPrefix(dpack.Message, getIRCConfigString("prefix"))
				debug("sending to \"" + message.Params[0])
				parseCommand(dpack)
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
