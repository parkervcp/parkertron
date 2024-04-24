package main

// discord configs
type discordBase struct {
	Bots []discordBot `json:"bots,omitempty"`
}

type discordBot struct {
	BotName string           `json:"bot_name,omitempty"`
	BotID   string           `json:"bot_id,omitempty"`
	Config  discordBotConfig `json:"config,omitempty"`
	Servers []discordServer  `json:"servers,omitempty"`
}

type discordBotConfig struct {
	Token  string        `json:"token,omitempty"`
	Game   string        `json:"game,omitempty"`
	DMResp responseArray `json:"dm_response,omitempty"`
}

type discordServer struct {
	ServerID    string              `json:"server_id,omitempty"`
	Config      discordServerConfig `json:"config,omitempty"`
	ChanGroups  []channelGroup      `json:"channel_groups,omitempty"`
	Permissions []permission        `json:"permissions,omitempty"`
	Filters     []filter            `json:"filters,omitempty"`
}

type discordServerConfig struct {
	Prefix   string          `json:"prefix,omitempty"`
	Clear    bool            `json:"clear_commands,omitempty"`
	WebHooks discordWebHooks `json:"web_hooks,omitempty"`
}

type discordWebHooks struct {
	Logs string `json:"logs,omitempty"`
}

type discordKickOnMention struct {
	Roles   []string      `json:"roles,omitempty"`
	Users   []string      `json:"users,omitempty"`
	Direct  responseArray `json:"dm,omitempty"`
	Channel responseArray `json:"channel,omitempty"`
	Kick    bool          `json:"kick,omitempty"`
}
