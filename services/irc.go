package services

import (
	"strings"
	"time"

	"github.com/husio/irc"
	"github.com/parkervcp/parkertron/config"
)

func initIRC(cfg *config.Config) {
	conn, err := irc.Connect(cfg.IRC.Address())
	if err != nil {
		return
	}

	conn.Send("USER %s %s * :%s", cfg.IRC.Ident, cfg.IRC.Host, cfg.IRC.RealName)
	conn.Send("NICK %s", cfg.IRC.Nickname)
	time.Sleep(time.Millisecond * 100)

	for {
		select {
		case <-shutdown:
			conn.Close()
			return
		default:
			msg, err := conn.ReadMessage()
			if err != nil {
				continue
			}

			switch msg.Command {
			case "PING":
				conn.Send("PONG %s", msg.Trailing)
			case "NOTICE":
				if strings.Contains(strings.ToLower(msg.Trailing), "this nickname is registered") {
					conn.Send("%s IDENTIFY %s %s", msg.Nick(), cfg.IRC.Nickname, cfg.IRC.Password)
				}
			case "PRIVMSG":
				// TODO
			}
		}
	}
}
