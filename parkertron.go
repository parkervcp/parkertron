package main

import (
	"bufio"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

var (
	startTime = time.Now

	// Log is a logrus logger
	Log *logrus.Logger

	//ServStat is the Service Status channel
	servStart   = make(chan string)
	shutdown    = make(chan string)
	servStopped = make(chan string)

	botConfig parkertron

	serviceStart = map[string]func(){
		"discord": startDiscordsBots,
		"irc":     startIRCBots,
	}

	serviceStop = map[string]func(){
		"discord": stopDiscordBots,
		"irc":     stopIRCBots,
	}

	// startup flag values
	verbose string
	logDir  string
	confDir string

	asciiArt = `
                      __             __
    ____  ____ ______/ /_____  _____/ /__________  ____
   / __ \/ __ '/ ___/ //_/ _ \/ ___/ __/ ___/ __ \/ __ \
  / /_/ / /_/ / /  / ,< /  __/ /  / /_/ /  / /_/ / / / /
 / .___/\__,_/_/  /_/|_|\___/_/   \__/_/   \____/_/ /_/
/_/ v.0.2.2`
)

type parkertron struct {
	Services []string       `json:"services"`
	Log      logConf        `json:"log"`
	Database databaseConfig `json:"database"`
}

type logConf struct {
	Level    string `json:"level"`
	Location string `json:"location"`
}

type databaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Database string `json:"database"`
}

func init() {
	flag.StringVarP(&verbose, "verbosity", "v", "info", "set the verbosity level for the bot {info,debug} (default is info)")
	flag.StringVarP(&logDir, "logdir", "l", "logs/", "set the log directory of the bot. (default is ./logs/)")
	flag.StringVarP(&confDir, "confdir", "c", "configs/", "set the config directory of the bot. (default is ./configs/)")
	flag.Parse()

	Log = newLogger(logDir, verbose)
	Log.Infof("logging online\n")

	initConfig(confDir)

	Log.Infof("%s\n\n", asciiArt)
}

func main() {
	for _, cr := range botConfig.Services {
		if service, ok := serviceStart[cr]; ok {
			Log.Debugf("running %s", runtime.FuncForPC(reflect.ValueOf(service).Pointer()).Name())
			go service()
		} else {
			Log.Errorf("unexpected array value: %q", cr)
		}
	}

	for range botConfig.Services {
		Log.Debugf("checking for servStart")
		<-servStart
	}

	go catchSig()
	go console()

	Log.Infof("Bot is now running. Send 'shutdown' or 'ctrl + c' to stop the bot .\n")

	<-shutdown

	for _, cr := range botConfig.Services {
		if service, ok := serviceStop[cr]; ok {
			Log.Debugf("running %s", runtime.FuncForPC(reflect.ValueOf(service).Pointer()).Name())
			go service()
		} else {
			Log.Errorf("unexpected array value: %q", cr)
		}
	}

	for range botConfig.Services {
		Log.Debugf("checking for servStopped")
		<-servStopped
	}
}

func console() {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			Log.Infof("cannot read from stdin: %v", err)
		}
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line == "shutdown" {
			Log.Infof("shutting down the bot.\n")
			Log.Infof("All services stopped\n")
			shutdown <- ""
			return
		}
	}
}

func catchSig() {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		os.Interrupt)
	<-sigc
	Log.Info("interupt caught")
	shutdown <- ""
}

func newLogger(logDir, level string) *logrus.Logger {
	if Log != nil {
		return Log
	}

	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  logDir + "info.log",
		logrus.ErrorLevel: logDir + "error.log",
	}

	Log = logrus.New()

	switch level {
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	Log.Hooks.Add(lfshook.NewHook(
		pathMap,
		&logrus.JSONFormatter{},
	))
	return Log
}
