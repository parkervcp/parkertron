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

type pattern struct {
	Pattern  string   `json:"pattern,omitempty"`
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
}

type mentions struct {
	Ping    responseArray `json:"ping,omitempty"`
	Mention responseArray `json:"mention,omitempty"`
}

type filter struct {
	Term   string   `json:"term,omitempty"`
	Reason []string `json:"reason,omitempty"`
}

type responseArray struct {
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
}

type parsing struct {
	Image parsingImageConfig `json:"image,omitempty"`
	Paste parsingPasteConfig `json:"paste,omitempty"`
}

type parsingConfig struct {
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	Format string `json:"format,omitempty"`
}

type parsingImageConfig struct {
	FileTypes []string        `json:"filetypes,omitempty"`
	Sites     []parsingConfig `json:"sites,omitempty"`
}

type parsingPasteConfig struct {
	Sites  []parsingConfig `json:"sites,omitempty"`
	Ignore []parsingConfig `json:"ignore,omitmepty"`
}

type channelGroup struct {
	ChannelIDs  []string             `json:"channels,omitempty"`
	Mentions    mentions             `json:"mentions,omitempty"`
	Commands    []command            `json:"commands,omitempty"`
	Keywords    []keyword            `json:"keywords,omitempty"`
	Regex       []pattern            `json:"regex,omitempty"`
	Parsing     parsing              `json:"parsing,omitempty"`
	Permissions []permission         `json:"permissions,omitempty"`
	KOM         discordKickOnMention `json:"kick_on_mention,omitempty"`
}
