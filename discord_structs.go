package main

// global discord config
type discord struct {
	Token   string          `json:"token,omitempty"`
	DMResp  string          `json:"dm_response,omitempty"`
	Game    string          `json:"game,omitempty"`
	Servers []discordServer `json:"server,omitempty"`
}

// Server config.
// Can support more than a single server.
// This contains the server info and other settings.
type discordServer struct {
	ServerID   string                `json:"server_id,omitempty"`
	Settings   discordServerSettings `json:"settings,omitempty"`
	ChanGroups []discordChannelGroup `json:"channel_groups,omitempty"`
}

// Server Settings
// This will be things like the prefix and permissions/channels
type discordServerSettings struct {
	Prefix        string                `json:"prefix,omitempty"`
	Channels      discordServerChannels `json:"server_channels,omitempty"`
	Webhooks      discordServerWebhooks `json:"webhooks,omitmepty"`
	Mentions      mentions              `json:"mentions,omitempty"`
	Permissions   []permGroups          `json:"permissions,omitempty"`
	Clearcommands bool                  `json:"clear_commands,omitempty"`
}

type discordServerChannels struct {
	Admin string `json:"admin,omitempty"`
	Log   string `json:"log,omitempty"`
}

type discordServerWebhooks struct {
	Log string `json:"log,omitempty"`
}

// channel config.
// Can support more than a single channel.
// This contains the channel info and other settings.
type discordChannelGroup struct {
	ChannelIDs []string  `json:"channels,omitempty"`
	UseGlobal  bool      `json:"use_global,omitempty"`
	Mentions   mentions  `json:"mentions,omitempty"`
	Commands   []command `json:"commands,omitempty"`
	Keywords   []keyword `json:"keywords,omitempty"`
	Parsing    parsing   `json:"parsing,omitempty"`
}
