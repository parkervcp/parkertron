package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	text "text/template"

	"gopkg.in/yaml.v3"
)

var (
	basepath string
	bot      Config
	mutex    sync.RWMutex
	servers  []ServerConfig
)

type reaction struct {
	name string `yaml:"name"`
	id   string `yaml:"id,omitempty"`
}

type Response struct {
	Message  string         `yaml:"message,omitempty"`
	Reaction *reaction      `yaml:"reaction,omitempty"`
	tmpl     *text.Template `yaml:"-"`
}

func (r *Response) MessageTemplate() *text.Template {
	return r.tmpl
}

func (r *Response) ReactionString() (string, bool) {
	if r.Reaction != nil {
		if r.Reaction.id != "" {
			return r.Reaction.name + ":" + r.Reaction.id, true
		}
		return r.Reaction.name, true
	}

	return "", false
}

type Command struct {
	Response        `yaml:",inline"`
	Name            string   `yaml:"name"`
	allowedChannels []string `yaml:"allowed_channels,omitempty"`
	allowedRoles    []string `yaml:"allowed_roles,omitempty"`
}

func (c *Command) HasChannel(id string) bool {
	if len(c.allowedChannels) == 0 {
		return true
	}

	for _, ch := range c.allowedChannels {
		if ch == id {
			return true
		}
	}

	return false
}

func (c *Command) HasRole(id string) bool {
	if len(c.allowedRoles) == 0 {
		return true
	}

	for _, ch := range c.allowedRoles {
		if ch == id {
			return true
		}
	}

	return false
}

type DMCommand struct {
	Response `yaml:",inline"`
	Name     string `yaml:"name"`
}

type Trigger struct {
	Response `yaml:",inline"`
	Pattern  string `yaml:"pattern"`
	Delete   bool   `yaml:"delete,omitempty"`
}

type ParseConfig struct {
	FileTypes []string `yaml:"allowed_filetypes"`
	Sites     []string `yaml:"allowed_sites"`
	Max       int      `yaml:"max"`
	OnParse   Response `yaml:"on_parse,omitempty"`
	OnMaxed   Response `yaml:"on_maxed,omitempty"`
}

func (p *ParseConfig) HasFileType(t string) bool {
	for _, f := range p.FileTypes {
		if f == t {
			return true
		}
	}

	return false
}

func (p *ParseConfig) HasSite(s string) bool {
	for _, u := range p.Sites {
		if u == s {
			return true
		}
	}

	return false
}

type ServerConfig struct {
	ID              string  `yaml:"server_id"`
	AllowedChannels []int64 `yaml:"allowed_channels"`
	AllowedRoles    []int64 `yaml:"allowed_roles"`

	OnMention Response    `yaml:"on_mention"`
	Commands  []*Command  `yaml:"commands"`
	Triggers  []*Trigger  `yaml:"triggers"`
	Parsing   ParseConfig `yaml:"parsing"`
}

type IRCConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Ident    string `yaml:"ident"`
	Password string `yaml:"password"`
	RealName string `yaml:"real_name"`
	Nickname string `yaml:"nickname"`
}

func (i *IRCConfig) Address() string {
	return i.Host + ":" + i.Port
}

type Config struct {
	DiscordToken     string      `yaml:"discord_token"`
	IRC              *IRCConfig  `yaml:"irc,omitempty"`
	Prefix           string      `yaml:"prefix"`
	Status           string      `yaml:"status"`
	DeleteInvocation bool        `yaml:"delete_invocation,omitempty"`
	DMCommands       []DMCommand `yaml:"dm_commands,omitempty"`
}

func init() {
	if basepath == "" {
		p := os.Getenv("PTRON_CONFIGS")
		if p == "" {
			cwd, _ := os.Getwd()
			p = filepath.Join(cwd, "configs")
		}
		basepath = p
	}
}

func LoadBot() (*Config, error) {
	path := filepath.Join(basepath, "bot.yml")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			path = filepath.Join(basepath, "bot.yaml")
			_, err = os.Stat(path)
		}

		if err != nil {
			return nil, err
		}
	}

	buf, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(buf, &bot)

	return &bot, err
}

func LoadServers() (int, error) {
	path := filepath.Join(basepath, "discord")
	dir, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	for _, e := range dir {
		if e.IsDir() {
			continue
		}

		if strings.HasSuffix(e.Name(), ".yml") || strings.HasSuffix(e.Name(), ".yaml") {
			buf, err := os.ReadFile(filepath.Join(path, e.Name()))
			if err != nil {
				return 0, err
			}

			var srv ServerConfig
			if err := yaml.Unmarshal(buf, &srv); err != nil {
				return 0, err
			}

			if srv.OnMention.Message != "" {
				tmpl, err := text.New(srv.ID).Parse(srv.OnMention.Message)
				if err == nil {
					srv.OnMention.tmpl = tmpl
				}
			}

			for _, c := range srv.Commands {
				if c.Message != "" {
					tmpl, err := text.New(c.Name).Parse(c.Message)
					if err != nil {
						continue
					}
					c.tmpl = tmpl
				}
			}

			for i, t := range srv.Triggers {
				if t.Message != "" {
					n := srv.ID + ".t." + strconv.Itoa(i)
					tmpl, err := text.New(n).Parse(t.Message)
					if err != nil {
						continue
					}
					t.tmpl = tmpl
				}
			}

			servers = append(servers, srv)
		}
	}

	return len(servers), nil
}

func GetServers() []ServerConfig {
	mutex.RLock()
	defer mutex.RUnlock()

	return servers
}

func GetServer(id string) *ServerConfig {
	mutex.RLock()
	defer mutex.RUnlock()

	for _, s := range servers {
		if s.ID == id {
			return &s
		}
	}

	return nil
}
