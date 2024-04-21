package services

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
	"github.com/parkervcp/parkertron/config"
	"mvdan.cc/xurls/v2"
)

const intents = dgo.IntentDirectMessages | dgo.IntentDirectMessageReactions | dgo.IntentGuildMessages | dgo.IntentGuildMessageReactions | dgo.IntentMessageContent

var (
	prefix string
	delete bool
	bot    *dgo.User
)

func initDiscord(cfg *config.Config) {
	dg, err := dgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		return
	}
	prefix = cfg.Prefix
	delete = cfg.DeleteInvocation

	dg.Identify.Intents = intents
	dg.AddHandler(func(dg *dgo.Session, r *dgo.Ready) {
		bot = r.User
		_ = dg.UpdateGameStatus(0, cfg.Status)
	})
	dg.AddHandler(handleMessage)

	if err := dg.Open(); err != nil {
		return
	}

	<-shutdown
	_ = dg.Close()
}

func handleMessage(dg *dgo.Session, evt *dgo.MessageCreate) {
	if evt.Author.Bot || evt.Content == "" {
		return
	}

	// ch, err := dg.Channel(evt.ChannelID)
	// if err != nil {
	// 	return
	// }

	// TODO: handle DM commands

	cfg := config.GetServer(evt.GuildID)
	if cfg == nil {
		return
	}

	if strings.HasPrefix(evt.Content, prefix) || strings.HasPrefix(evt.Content, bot.Mention()) {
		go handleGuildCommand(dg, evt, cfg)
	} else {
		go handleGuildMessage(dg, evt, cfg)
	}
}

// func handleDMCommand(dg *dgo.Session, evt *dgo.MessageCreate) {}

func handleGuildCommand(dg *dgo.Session, evt *dgo.MessageCreate, cfg *config.ServerConfig) {
	if strings.HasPrefix(evt.Content, prefix) {
		evt.Content = evt.Content[len(prefix):]
	} else {
		evt.Content = strings.TrimSpace(evt.Content[len(bot.Mention()):])
		if len(evt.Content) == 0 {
			respond(dg, evt, &cfg.OnMention, false)
			return
		}
	}

	args := strings.Split(evt.Content, " ")
	var cmd *config.Command

	for _, c := range cfg.Commands {
		if c.Name == args[0] {
			cmd = c
			break
		}
	}

	if cmd == nil {
		return
	}

	if !cmd.HasChannel(evt.ChannelID) {
		return
	}
	// TODO: handle roles

	if delete {
		_ = dg.ChannelMessageDelete(evt.ChannelID, evt.ID)
	}

	respond(dg, evt, &cmd.Response, delete)
}

func handleGuildMessage(dg *dgo.Session, evt *dgo.MessageCreate, cfg *config.ServerConfig) {
	for _, t := range cfg.Triggers {
		if ok, _ := regexp.MatchString(t.Pattern, evt.Content); ok {
			respond(dg, evt, &t.Response, t.Delete)
			return
		}
	}

	var urls []*url.URL
	for _, u := range xurls.Relaxed().FindAllString(evt.Content, -1) {
		p, err := url.Parse(u)
		if err != nil {
			continue
		}

		if cfg.Parsing.HasSite(p.Hostname()) {
			urls = append(urls, p)
		}
	}

	if len(urls) > cfg.Parsing.Max {
		respond(dg, evt, &cfg.Parsing.OnMaxed, false)
		return
	}

	respond(dg, evt, &cfg.Parsing.OnParse, false)
	// TODO: connect to parsing functions
}

func respond(dg *dgo.Session, evt *dgo.MessageCreate, res *config.Response, delete bool) {
	if delete {
		_ = dg.ChannelMessageDelete(evt.ChannelID, evt.ID)

		if t := res.MessageTemplate(); t != nil {
			buf := bytes.Buffer{}
			if err := t.Execute(&buf, evt.Author); err != nil {
				return
			}

			_, err := dg.ChannelMessageSend(evt.ChannelID, buf.String())
			if err != nil {
				_ = err
			}
		}
		return
	}

	if r, ok := res.ReactionString(); ok {
		if err := dg.MessageReactionAdd(evt.ChannelID, evt.ID, r); err != nil {
			_ = err
		}
	}

	if t := res.MessageTemplate(); t != nil {
		buf := bytes.Buffer{}
		if err := t.Execute(&buf, evt.Author); err != nil {
			return
		}

		_, err := dg.ChannelMessageSendComplex(evt.ChannelID, &dgo.MessageSend{
			AllowedMentions: &dgo.MessageAllowedMentions{RepliedUser: false},
			Content:         buf.String(),
			Reference:       evt.Reference(),
		})
		if err != nil {
			_ = err
		}
	}
}
