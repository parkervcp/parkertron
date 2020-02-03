package main

import (
	"strings"
	"time"

	hirc "github.com/husio/irc"
)

var (
	stopIRC = make(map[string](chan string))

	ircGlobal irc

	ircLoad = make(chan string)
)

//ircMessageHandler the IRC listener that manages inbound messaging
func ircMessageHandler(conn hirc.Conn, ircConfig ircBot) {
	message, err := conn.ReadMessage()
	if err != nil {
		Log.Errorf("cannot read message: ", err)
		return
	}

	// Log.Debugf("started handle")
	Log.Debug("irc inbound " + message.String())

	// keep alive messaging
	if message.Command == "PING" {
		conn.Send("PONG " + message.Trailing)
		Log.Debug("PONG Sent")
		return
	}

	// for authentication
	if message.Command == "NOTICE" {
		if strings.Contains(strings.ToLower(message.Trailing), "this nickname is registered") {
			conn.Send("%s IDENTIFY %s %s", message.Nick(), ircConfig.Config.Server.Nickname, ircConfig.Config.Server.Password)
		}
		return
	}

	// message handling
	if message.Command == "PRIVMSG" {
		Log.Debug("message.Params[0]: " + message.Params[0]) // channel
		Log.Debug("message.Nick(): " + message.Nick())       // user
		Log.Debug("message.Trailing: " + message.Trailing)   // actual message

		var blacklistedUsers string
		var channelCommands []command
		var channelKeywords []keyword
		var channelParsing parsing

		for gid, group := range ircConfig.ChanGroups {
			for _, channel := range group.ChannelIDs {
				if message.Params[0] == channel {
					for pid, perms := range ircConfig.ChanGroups[gid].Permissions {
						if perms.Blacklisted {
							for _, user := range ircConfig.ChanGroups[gid].Permissions[pid].Users {
								blacklistedUsers = blacklistedUsers + ", " + user
							}
						}
					}
					channelCommands = group.Commands
					channelKeywords = group.Keywords
				}
			}
		}

		// if the user nickname matches bot or blacklisted.
		if message.Nick() == ircConfig.Config.Server.Nickname || strings.Contains(blacklistedUsers, message.Nick()) {
			if message.Nick() == ircConfig.Config.Server.Nickname {
				Log.Debug("User is the bot and being ignored.")
				return
			}
			if strings.Contains(blacklistedUsers, message.Nick()) {
				Log.Debug("User is blacklisted")
				return
			}
		}

		// if bot is DM'd
		if message.Params[0] == ircConfig.Config.Server.Nickname {
			Log.Debug("This was a DM")
			sendIRCMessage(conn, message.Params[0], message.Nick(), ircConfig.Config.Prefix, ircConfig.Config.DMResp)
			return
		}

		//
		// Message Handling
		//
		if message.Trailing != "" {
			Log.Debug("Message Content: " + message.Trailing)

			if !strings.HasPrefix(message.Trailing, ircConfig.Config.Prefix) {
				Log.Debug("sending to \"" + message.Params[0])
				parseKeyword(message.Trailing, ircConfig.BotName, []string{}, channelKeywords, channelParsing)
			} else {
				Log.Debug("sending to \"" + message.Params[0])
				parseCommand(strings.TrimPrefix(message.Trailing, ircConfig.Config.Prefix), ircConfig.BotName, channelCommands)
			}
			return
		}
		// Log.Debug(message.Raw)
	}
	return
}

// kick irc user
func kickIRCUser() {

}

// ban irc user
func banIRCUser() {

}

//sendIRCMessage function to send messages separate of the listener
func sendIRCMessage(conn hirc.Conn, channelName string, user string, prefix string, responseArray []string) {
	// send nothing if there is nothing in the array
	if len(responseArray) == 0 {
		return
	}

	// send a line per item in the array.
	for _, response := range responseArray {
		Log.Debugf("line sent: " + response)
		response = strings.Replace(response, "&user&", user, -1)
		response = strings.Replace(response, "&prefix&", prefix, -1)
		conn.Send("PRIVMSG " + "#" + channelName + " :" + response)
		time.Sleep(time.Millisecond * 300)
	}

	// log the message that was sent in debug mode.
	Log.Debugf("IRC Message Sent: %s", responseArray)
}

// service handling
// start all the bots
func startIRCBots() {
	Log.Infof("Starting IRC server connections\n")
	// range over the bots available to start
	for _, bot := range ircGlobal.Bots {
		Log.Infof("Connecting to %s\n", bot.BotName)

		// spin up a channel to tell the bot to shutdown later
		stopIRC[bot.BotName] = make(chan string)

		// start the bot
		go startIRCConnection(bot)
		// wait on bot being able to start.
		<-ircLoad
	}

	Log.Infof("irc service started\n")
	// inform main process that the bot is started
	servStart <- "irc_online"
}

// when a shutdown is sent close out services properly
func stopIRCBots() {
	Log.Infof("stopping irc connections")
	// loop through bots and send shutdowns
	for _, bot := range ircGlobal.Bots {
		Log.Infof("stopping %s", bot.BotName)
		// send stop to bot
		stopIRC[bot.BotName] <- ""
		// wait for bot to send a stop back
		<-stopIRC[bot.BotName]
		// close channel
		close(stopIRC[bot.BotName])
		Log.Infof("stopped %s", bot.BotName)
	}

	Log.Infof("irc connections stopped")
	// return shutdown signal on channel
	servStopped <- "irc_stopped"
}

// start connections to irc
func startIRCConnection(ircConfig ircBot) {
	host := ircConfig.Config.Server.Address + ":" + ircConfig.Config.Server.Port
	Log.Debugf("Connecting on %s\n", host)

	// Connect to the server
	conn, err := hirc.Connect(host)
	if err != nil {
		Log.Errorf("cannot connect to %s: %s\n", host, err)
	}

	Log.Debugf("Connected to %s\n", host)

	// send user info
	conn.Send("USER %s %s * :"+ircConfig.Config.Server.RealName, ircConfig.Config.Server.Ident, host)
	conn.Send("NICK %s", ircConfig.Config.Server.Nickname)

	time.Sleep(time.Millisecond * 100)

	ircLoad <- ""

	for _, group := range ircConfig.ChanGroups {
		for _, channel := range group.ChannelIDs {
			Log.Debugf("joining %s", channel)
			if !strings.HasPrefix(channel, "#") {
				channel = "#" + channel
			}
			conn.Send("JOIN %s", channel)
		}
	}

	for {
		// listen for stop on channel
		select {
		case <-stopIRC[ircConfig.BotName]:
			Log.Debugf("closing channel for %s", ircConfig.BotName)
			conn.Close()
			stopIRC[ircConfig.BotName] <- ""
			Log.Debugf("%s channel closed", ircConfig.BotName)
			return
		default:
			ircMessageHandler(*conn, ircConfig)
		}
	}
}
