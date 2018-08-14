package main

import (
	"os"
	"time"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"

	"github.com/kz/discordrus"
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
	log.AddHook(discordrus.NewHook(
		// Use environment variable for security reasons
		getDiscordConfigString("webhook.log"),
		// Set minimum level to DebugLevel to receive all log entries
		log.DebugLevel,
		&discordrus.Opts{
			Username:           "Test Username",
			Author:             "",                         // Setting this to a non-empty string adds the author text to the message header
			DisableTimestamp:   false,                      // Setting this to true will disable timestamps from appearing in the footer
			TimestampFormat:    "Jan 2 15:04:05.00000 MST", // The timestamp takes this format; if it is unset, it will take logrus' default format
			TimestampLocale:    nil,                        // The timestamp uses this locale; if it is unset, it will use time.Local
			EnableCustomColors: true,                       // If set to true, the below CustomLevelColors will apply
			CustomLevelColors: &discordrus.LevelColors{
				Debug: 10170623,
				Info:  3581519,
				Warn:  14327864,
				Error: 13631488,
				Panic: 13631488,
				Fatal: 13631488,
			},
			DisableInlineFields: false, // If set to true, fields will not appear in columns ("inline")
		},
	))
	info("Discord webhook logging online")
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

func auditKick() {

}

func auditBan() {

}
