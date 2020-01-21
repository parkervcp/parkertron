package main

type irc struct {
	Server []ircServer `json:"server,omitempty"`
}

type ircServer struct {
	Name       string             `json:"name,omitempty"`
	Config     ircServerConfig    `json:"config,omitempty"`
	DmResponse []string           `json:"dm_response,omitempty"`
	Channels   []ircChannelConfig `json:"channel_config,omitempty"`
}

type ircServerConfig struct {
	Connection ircServerConnectionConfig `json:"connection,omitempty"`
	User       ircServerUserConfig       `json:"user,omitempty"`
}

type ircServerConnectionConfig struct {
	Address    string `json:"address,omitempty"`
	Port       string `json:"port,omitempty"`
	SslEnabled bool   `json:"ssl_enabled,omitempty"`
}

type ircServerUserConfig struct {
	Identity string `json:"identity,omitempty"`
	Password string `json:"password,omitempty"`
	Nick     string `json:"nick,omitempty"`
	Real     string `json:"real,omitempty"`
	Email    string `json:"email,omitempty,omitempty"`
}

type ircChannelConfig struct {
	Name     string             `json:"name,omitempty"`
	Names    []string           `json:"names,omitempty"`
	Settings ircChannelSettings `json:"settings,omitempty"`
	Mentions mentions           `json:"mentions,omitempty"`
	Commands []command          `json:"commands,omitempty"`
	Keywords []keyword          `json:"keywords,omitempty"`
	Parsing  parsing            `json:"parsing,omitempty"`
}

type ircChannelSettings struct {
	UseGlobal   bool       `json:"use_global,omitempty"`
	Prefix      string     `json:"prefix,omitempty"`
	Permissions permGroups `json:"permissions,omitempty"`
}
