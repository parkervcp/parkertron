package services

import "github.com/parkervcp/parkertron/config"

var shutdown = make(chan struct{}, 1)

func Start(cfg *config.Config) {
	go initDiscord(cfg)
	go initIRC(cfg)
}

func Stop() {
	shutdown <- struct{}{}
}
