package main

// irc configs
type irc struct {
	Bots []ircBot `json:"bots,omitempty"`
}

type ircBot struct {
	BotName    string         `json:"bot_name,omitempty"`
	Config     ircBotConfig   `json:"config,omitempty"`
	ChanGroups []channelGroup `json:"channel_groups,omitempty"`
}

type ircBotConfig struct {
	Server ircServerConfig `json:"server,omitempty"`
	DMResp responseArray   `json:"dm_response,omitempty"`
	Prefix string          `json:"prefix,omitempty"`
}

type ircServerConfig struct {
	Address   string `json:"address,omitempty"`
	Port      int    `json:"port,omitempty"`
	SSLEnable bool   `json:"ssl,omitempty"`
	Ident     string `json:"ident,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"password,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	RealName  string `json:"real_name,omitempty"`
}
