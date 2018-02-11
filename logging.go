package main

import (
	"os"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

//Log is a here for the NewLogger
var (
	Log     *log.Logger
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
}

//NewLogger is the filesystem hook for logrus
func NewLogger() *log.Logger {
	if Log != nil {
		return Log
	}

	pathMap := lfshook.PathMap{
		log.InfoLevel:  "logs/info.log",
		log.ErrorLevel: "logs/error.log",
	}

	Log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&log.JSONFormatter{},
	))
	return Log
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
