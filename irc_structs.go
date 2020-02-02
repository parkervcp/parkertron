package main

// irc configs
type irc struct {
	Bots []ircBot `json:"bots,omitempty"`
}

type ircBot struct {
	BotName    string             `json:"bot_name,omitempty"`
	Config     ircBotConfig       `json:"config,omitempty"`
	ChanGroups []ircChannelGroups `json:"channel_groups,omitempty"`
}

type ircBotConfig struct {
	Server ircServerConfig `json:"server,omitempty"`
	DMResp []string        `json:"dm_response,omitempty"`
	Prefix string          `json:"prefix,omitempty"`
}

type ircServerConfig struct {
	Address   string `json:"address,omitempty"`
	Port      string `json:"port,omitempty"`
	SSLEnable bool   `json:"ssl,omitempty"`
	Ident     string `json:"ident,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	RealName  string `json:"real_name,omitempty"`
}

type ircChannelGroups struct {
	ChannelIDs  []string     `json:"channels,omitempty"`
	Permissions []permission `json:"permissions,omitempty"`
	Mentions    mentions     `json:"mentions,omitempty"`
	Commands    []command    `json:"commands,omitempty"`
	Keywords    []keyword    `json:"keywords,omitempty"`
	Parsing     parsing      `json:"parsing,omitempty"`
}
