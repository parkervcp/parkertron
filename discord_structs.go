package main

// discord configs
type discord struct {
	Bots []discordBot `json:"bots,omitempty"`
}

type discordBot struct {
	BotName string           `json:"bot_name,omitempty"`
	BotID   string           `json:"bot_id,omitempty"`
	Config  discordBotConfig `json:"config,omitempty"`
	Servers []discordServer  `json:"servers,omitempty"`
}

type discordBotConfig struct {
	Token  string   `json:"token,omitempty"`
	Game   string   `json:"game,omitempty"`
	DMResp []string `json:"dm_response,omitempty"`
}

type discordServer struct {
	ServerID    string                 `json:"server_id,omitempty"`
	Config      discordServerConfig    `json:"config,omitempty"`
	ChanGroups  []discordChannelGroups `json:"channel_groups,omitempty"`
	Permissions []permission           `json:"permissions,omitempty"`
}

type discordServerConfig struct {
	Prefix   string          `json:"prefix,omitempty"`
	Clear    bool            `json:"clear_commands,omitempty"`
	Webhooks discordWebhooks `json:"webhooks,omitempty"`
}

type discordWebhooks struct {
	Logs string `json:"logs,omitempty"`
}

type discordChannelGroups struct {
	ChannelIDs []string             `json:"channels,omitempty"`
	Mentions   mentions             `json:"mentions,omitempty"`
	Commands   []command            `json:"commands,omitempty"`
	Keywords   []keyword            `json:"keywords,omitempty"`
	Parsing    parsing              `json:"parsing,omitempty"`
	KOM        discordKickOnMention `json:"kick_on_mention,omitempty"`
}

type discordKickOnMention struct {
	Roles   []string `json:"roles,omitempty"`
	Users   []string `json:"users,omitempty"`
	Direct  response `json:"dm,omitempty"`
	Channel response `json:"channel,omitempty"`
}
