package main

import (
	"strings"
	"time"

	hirc "github.com/husio/irc"
)

var (
	stopIRC = make(chan string)

	ircConfig irc

	address = ircConfig.Server + ":" + getIRCConfigString("server.port")

	c, err = hirc.Connect(address)
)

func checkIRCMessage(conn hirc.Conn, msgContent string, botNick string, channelName string, userNick string, channelConf ircChannelConfig) bool {
	Log.Debugf("prefix is %s", channelConf.Settings.Prefix)
	// Channel specific matching
	if msgContent == "@"+botNick {
		sendIRCMessage(conn, channelName, userNick, channelConf.Settings.Prefix, channelConf.Mentions.Ping)
		return true
	}

	if strings.Contains(msgContent, "@"+botNick) {
		sendIRCMessage(conn, channelName, userNick, channelConf.Settings.Prefix, channelConf.Mentions.Mention)
		return true
	}

	//
	// Message Handling
	//
	if msgContent != "" {
		Log.Debugf("Message Content: " + msgContent)
		// determine if the mesage is a command or a keywordsparse
		if strings.HasPrefix(msgContent, channelConf.Settings.Prefix) {
			Log.Debugf("command recieved %s", strings.TrimPrefix(msgContent, channelConf.Settings.Prefix))
			response := parseCommand(channelConf.Commands, strings.TrimPrefix(msgContent, channelConf.Settings.Prefix))
			sendIRCMessage(conn, channelName, userNick, channelConf.Settings.Prefix, response)
			return true
		}
		Log.Debugf("keyword recieved %s", msgContent)
		response, _ := parseKeyword(channelConf.Keywords, channelConf.Parsing, msgContent)
		sendIRCMessage(conn, channelName, userNick, channelConf.Settings.Prefix, response)
		return true
	}
	return false
}

// loops through and starts the connections and listeners for each server.
func handleIRC(serverName string) {
	var serverConf ircServer
	// This is the address of the irc server and the port combined to make it easier to input later
	for _, server := range ircC.Server {
		if server.Name == serverName {
			serverConf = server
		}
	}

	host := serverConf.Config.Connection.Address + ":" + serverConf.Config.Connection.Port
	Log.Debugf("Connecting on %s\n", host)

	// Connect to the server
	conn, err := hirc.Connect(host)
	if err != nil {
		Log.Errorf("cannot connect to %s: %s\n", host, err)
	}

	Log.Debugf("Connected to %s\n", host)

	// send user info
	conn.Send("USER %s %s * :"+serverConf.Config.User.Real, serverConf.Config.User.Identitiy, host)
	conn.Send("NICK %s", serverConf.Config.User.Nick)

	time.Sleep(time.Millisecond * 100)

	for {
		handleIRCMessages(*conn, serverConf)
		select {
		case x, ok := <-stopIRC:
			Log.Debugf("Disconnecting from %s", x)
			if ok && x == serverName {
				conn.Close()
				stopIRC <- ""
				return
			}
		default:
		}
	}
}

//ircMessageHandler the IRC listener that manages inbound messaging
func ircMessageHandler() {
	message, err := c.ReadMessage()
	if err != nil {
		Log.Fatal("cannot read message: ", err)
		return
	}

	Log.Debug("irc inbound " + message.String())

	// keep alive messaging
	if message.Command == "PING" {
		c.Send("PONG " + message.Trailing)
		Log.Debug("PONG Sent")
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
		Log.Debug("message.Params[0]: " + message.Params[0])
		Log.Debug("message.Nick(): " + message.Nick())
		Log.Debug("message.Trailing: " + message.Trailing)

		// if the user nickname matches bot or blacklisted.
		if message.Nick() == dpack.BotID || strings.Contains(getIRCBlacklist(), dpack.AuthorID) {
			if message.Nick() == dpack.BotID {
				Log.Debug("User is the bot and being ignored.")
				return
			}
			if strings.Contains(getIRCBlacklist(), dpack.AuthorID) {
				Log.Debug("User is blacklisted")
				return
			}
		}

		// if bot is DM'd
		if message.Params[0] == getIRCConfigString("nick") {
			Log.Debug("This was a DM")
			dpack.Response = getDiscordConfigString("direct.response")
			sendIRCMessage(message.Nick(), getIRCConfigString("direct.response"))
			return
		}

		//
		// Message Handling
		//
		if dpack.Message != "" {
			Log.Debug("Message Content: " + dpack.Message)

			if !strings.HasPrefix(dpack.Message, getIRCConfigString("prefix")) {
				Log.Debug("sending to \"" + message.Params[0])
				parseKeyword(dpack)
			} else {
				dpack.Message = strings.TrimPrefix(dpack.Message, getIRCConfigString("prefix"))
				Log.Debug("sending to \"" + message.Params[0])
				parseCommand(dpack)
			}
			return
		}
		Log.Debug(message.Raw)
	}
}

//sendIRCMessage function to send messages separate of the listener
func sendIRCMessage(conn hirc.Conn, channelName string, user string, prefix string, responseArray []string) {
	if len(responseArray) == 0 {
		return
	}

	for _, response = range responseArray {
		Log.Debugf("line sent: " + response)
		response = strings.Replace(response, "&user&", user, -1)
		response = strings.Replace(response, "&prefix&", prefix, -1)
		conn.Send("PRIVMSG " + "#" + channelName + " :" + response)
	}
	Log.Debugf("IRC Message Sent: %s", responseArray)
}

// service handling
// start connections to irc
func startIRCConnection() {
	Log.Infof("Starting IRC server connections\n")
	for _, server := range ircC.Server {
		Log.Infof("Connecting to %s\n", server.Name)
		go handleIRC(server.Name)
	}

	Log.Debugf("irc service started\n")

	servStat <- "irc_online"
}

// stop connections to irc
func stopIRCConnection() {
	for _, server := range ircC.Server {
		Log.Infof("stopping irc connection on %s", server.Name)
		stopIRC <- server.Name
		<-stopIRC
		Log.Infof("irc connection on %s stopped", server.Name)
	}
	shutdown <- ""
}

//registration services
func registerIRCUserFreenode(conn hirc.Conn, channelName string) {

}

func registerIRCUserQuakenet(conn hirc.Conn, channelName string) {

}
