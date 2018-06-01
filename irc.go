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

	if message.Command == "PRIVMSG" {
		input := message.Trailing

		if message.Nick() == getIRCConfigString("nick") || strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
			if message.Nick() == getIRCConfigString("nick") {
				writeLog("debug", "User is the bot and being ignored.", nil)
				return
			}
			if strings.Contains(getIRCGroupMembers("blacklist"), message.Params[0]) {
				writeLog("debug", "User is blacklisted", nil)
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
			writeLog("debug", "Message Content: "+input+"\n", nil)

			if strings.HasPrefix(input, getIRCConfigString("prefix")) == false {
				writeLog("debug", "sending to \""+message.Params[0], nil)
				parseKeyword("irc", message.Params[0], input)
				return
			} else if strings.HasPrefix(input, getIRCConfigString("prefix")) == true {
				input := strings.TrimPrefix(input, getIRCConfigString("prefix"))
				writeLog("debug", "sending to \""+message.Params[0], nil)
				parseCommand("irc", message.Params[0], input)
				return
			}
			return
		}
		writeLog("debug", message.Raw, nil)
	}
}

//sendIRCMessage function to send messages separate of the listener
func sendIRCMessage(ChannelID string, response string) {
	response = strings.Replace(response, "&prefix&", getIRCConfigString("prefix"), -1)
	multiresp := strings.Split(response, "\n")

	writeLog("debug", "IRC Message Sent:", nil)

	for x := range multiresp {
		writeLog("debug", "line sent: "+multiresp[x], nil)
		c.Send("PRIVMSG " + ChannelID + " :" + multiresp[x])
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
