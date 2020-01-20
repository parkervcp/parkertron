package main

import (
	"os"
	"time"

	"github.com/rifflock/lfshook"
	Log "github.com/sirupsen/logrus"
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
			Log.Error("failed to move latest logs.", err)
			return
		}
	}

	if _, err := os.Stat("./logs/debug.log"); err == nil {
		err := os.Rename("./logs/debug.log", "./logs/debug-"+time.Now().UTC().Format("2006-01-02 15:04")+".log")

		if err != nil {
			Log.Error("failed to move debug logs.", err)
			return
		}
	}
	Log.Info("Bot logging online")
}

func setLogLevel(level string) {
	if level == "debug" {
		Log.SetLevel(Log.DebugLevel)
		Log.Debug("log level set to debug")
	} else if level == "info" {
		Log.SetLevel(Log.InfoLevel)
		Log.Debug("log level set to info")
	}

	pathMap := lfshook.PathMap{
		Log.InfoLevel:  "logs/latest.log",
		Log.ErrorLevel: "logs/latest.log",
		Log.DebugLevel: "logs/debug.log",
	}
	Log.AddHook(lfshook.NewHook(
		pathMap,
		&Log.JSONFormatter{},
	))

}

func auditKick() {

}

func auditBan() {

}
