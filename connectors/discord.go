package connectors

import "time"

var (
	nextSend = time.Now()
)

func discordHandler() {
	// manage connection to discord
}

func messageGet() {
	// handle inbound messaging
}

func messageSend() {
	// handle outbound messaging
}

func dPerms() {
	// handle discord user perms
}
