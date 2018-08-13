package main

import (
	"math/rand"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand

func discordChannelFilter(req string) bool {
	if getDiscordConfigBool("channels.filter") == true {
		if strings.Contains(getDiscordChannels(), req) {
			return true
		}
		if getDiscordKOMChannel(req) {
			return true
		}
		return false
	}
	return true
}

func discordPermissioncheck(authorID string) (bool, string) {
	for _, y := range getDiscordGroupMembers("admin") {
		if authorID == y {
			debug("User " + authorID + "is an admin")
			return true, "admin"
		}
	}
	for _, y := range getDiscordGroupMembers("mod") {
		if authorID == y {
			debug("User " + authorID + "is a mod")
			return true, "mod"
		}
	}
	debug("User " + authorID + " is not in a perms group")
	return false, ""
}

func discordMessageHandler(message string, channel string, messageID string, authorID string, perms bool, group string) {
	// If the string doesn't have the prefix parse as text, if it does parse as a command.
	if !strings.HasPrefix(message, getDiscordConfigString("prefix")) {
		if discordChannelFilter(channel) {
			debug("No prefix was found parsing for keywords.")
			parseKeyword("discord", channel, message)
		}
	} else {
		// if there is a prefix check permissions on the user and run commands per group.
		message = strings.TrimPrefix(message, getDiscordConfigString("prefix"))
		if perms {
			if group == "admin" {
				parseAdminCommand("discord", channel, authorID, message)
			}
			if group == "mod" {

			}
		}
		// parse commands for matches
		debug("Prefix was found parsing for commands.")
		message = strings.TrimPrefix(message, getDiscordConfigString("prefix"))
		parseCommand("discord", channel, authorID, message)
		// remove previous commands if discord.command.remove is true
		if getDiscordConfigBool("command.remove") {
			if getCommandStatus(message) {
				deleteDiscordMessage(channel, messageID)
				debug("Cleared command message.")
			}
			if strings.HasPrefix(message, "list") || strings.HasPrefix(message, "ggl") {
				deleteDiscordMessage(channel, messageID)
				debug("Cleared command message.")
			}
		}
	}
}

func discordAttachmentHandler(attachments []string, channelID string) {
	for _, y := range attachments {
		debug("Sending attachment links to image parser")
		parseKeyword("discord", channelID, y)
	}
}

func discordImageRandGen() string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
