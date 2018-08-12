package main

import (
	"os"
	"time"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

//Log is a here for the NewLogger
var (
	level   string
	message string
)

func setupLogger() {
	if _, err := os.Stat("logs/"); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir("logs/", 0755)
		} else {
			// other error
		}
	}

	if _, err := os.Stat("./logs/latest.log"); err == nil {
		err := os.Rename("./logs/latest.log", "./logs/"+time.Now().UTC().Format("2006-01-02 15:04")+".log")

		if err != nil {
			errmsg("failed to move latest logs.", err)
			return
		}
	}

	if _, err := os.Stat("./logs/debug.log"); err == nil {
		err := os.Rename("./logs/debug.log", "./logs/debug-"+time.Now().UTC().Format("2006-01-02 15:04")+".log")

		if err != nil {
			errmsg("failed to move debug logs.", err)
			return
		}
	}
	info("Bot logging online")
}

func setLogLevel(level string) {
	if level == "debug" {
		log.SetLevel(log.DebugLevel)
		debug("log level set to debug")
	} else if level == "info" {
		log.SetLevel(log.InfoLevel)
		debug("log level set to info")
	}

	pathMap := lfshook.PathMap{
		log.InfoLevel:  "logs/latest.log",
		log.ErrorLevel: "logs/latest.log",
		log.DebugLevel: "logs/debug.log",
	}
	log.AddHook(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))

}

func info(message string) {
	log.Info(message)
}

func debug(message string) {
	log.Debug(message)
}

func superdebug(message string) {
	if getBotConfigString("log.level") == "superdebug" {
		log.Debug(message)
	}
}

func errmsg(message string, err error) {
	log.Error(message)
	log.Error(err)
}

func warn(message string) {
	log.Warn(message)
}

func fatal(message string, err error) {
	log.Fatal(message)
	log.Fatal(err)
}

func panic(message string) {
	log.Panic(message)
}
