package main

import (
	"os"

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
}

func setLogLevel(level string) {
	if level == "debug" {
		log.SetLevel(log.DebugLevel)
		writeLog("debug", "log level set to debug", nil)
	} else if level == "info" {
		log.SetLevel(log.InfoLevel)
		writeLog("debug", "log level set to info", nil)
	}

	if level == "debug" {
		pathMap := lfshook.PathMap{
			log.InfoLevel:  "logs/info.log",
			log.ErrorLevel: "logs/error.log",
			log.DebugLevel: "logs/debug.log",
		}
		log.AddHook(lfshook.NewHook(
			pathMap,
			&log.JSONFormatter{},
		))
	} else {
		pathMap := lfshook.PathMap{
			log.InfoLevel:  "logs/info.log",
			log.ErrorLevel: "logs/error.log",
		}
		log.AddHook(lfshook.NewHook(
			pathMap,
			&log.JSONFormatter{},
		))
	}
}

func writeLog(level string, message string, err error) {
	switch {
	case level == "debug":
		log.Debug(message)
	case level == "info":
		log.Info(message)
	case level == "warn":
		log.Warn(message)
	case level == "error":
		log.Error(message)
	case level == "fatal":
		log.Fatal(message)
	case level == "panic":
		log.Panic(message)
	}

	if err != nil {
		log.Fatal(err)
	}

}
