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

func discordMessageHandler(dpack DataPackage) {
	// If the string doesn't have the prefix parse as text, if it does parse as a command.
	if !strings.HasPrefix(message, getDiscordConfigString("prefix")) {
		if discordChannelFilter(dpack.ChannelID) {
			debug("No prefix was found parsing for keywords.")
			parseKeyword(dpack)
		}
	} else {
		// if there is a prefix check permissions on the user and run commands per group.
		message = strings.TrimPrefix(message, getDiscordConfigString("prefix"))
		if dpack.Perms {
			if dpack.Group == "admin" {
				parseAdminCommand(dpack)
				parseModCommand(dpack)
			}
			if dpack.Group == "mod" {
				parseModCommand(dpack)
			}
		}
		// parse commands for matches
		debug("Prefix was found parsing for commands.")
		message = strings.TrimPrefix(message, getDiscordConfigString("prefix"))
		parseCommand(dpack)
		// remove previous commands if discord.command.remove is true
		if getDiscordConfigBool("command.remove") {
			if getCommandStatus(message) {
				deleteDiscordMessage(dpack)
				debug("Cleared command message.")
			}
			if strings.HasPrefix(message, "list") || strings.HasPrefix(message, "ggl") {
				deleteDiscordMessage(dpack)
				debug("Cleared command message.")
			}
		}
	}
}

func discordImageRandGen() string {
	b := make([]byte, 12)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
