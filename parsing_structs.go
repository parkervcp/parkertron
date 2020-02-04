package main

// generic structs
type permission struct {
	Group       string   `json:"group,omitempty"`
	Users       []string `json:"users,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Commands    []string `json:"commands,omitempty"`
	Blacklisted bool     `json:"blacklisted,omitempty"`
}

type command struct {
	Command  string   `json:"command,omitempty"`
	Response []string `json:"response,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
}

type keyword struct {
	Keyword  string   `json:"keyword,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
	Exact    bool     `json:"exact,omitempty"`
}

type mentions struct {
	Ping    responseArray `json:"ping,omitempty"`
	Mention responseArray `json:"mention,omitempty"`
}

type responseArray struct {
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
}

type parsing struct {
	Image parsingImageConfig   `json:"image,omitempty"`
	Paste []parsingPasteConfig `json:"paste,omitempty"`
}

type parsingPasteConfig struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format,omitempty"`
}

type parsingImageConfig struct {
	Filetypes []string `json:"filetypes,omitempty"`
}

type channelGroups struct {
	ChannelIDs  []string             `json:"channels,omitempty"`
	Mentions    mentions             `json:"mentions,omitempty"`
	Commands    []command            `json:"commands,omitempty"`
	Keywords    []keyword            `json:"keywords,omitempty"`
	Parsing     parsing              `json:"parsing,omitempty"`
	Permissions []permission         `json:"permissions,omitempty"`
	KOM         discordKickOnMention `json:"kick_on_mention,omitempty"`
}
