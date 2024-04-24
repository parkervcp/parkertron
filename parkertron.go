package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
	"github.com/syfaro/haste-client"
)

var (
	//startTime = time.Now

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
	conf    string
	diag    bool
)

const (
	version  = "v.0.4.0"
	asciiArt = `
                      __             __
    ____  ____ ______/ /_____  _____/ /__________  ____
   / __ \/ __ '/ ___/ //_/ _ \/ ___/ __/ ___/ __ \/ __ \
  / /_/ / /_/ / /  / ,< /  __/ /  / /_/ /  / /_/ / / / /
 / .___/\__,_/_/  /_/|_|\___/_/   \__/_/   \____/_/ /_/
/_/`
)

type parkertron struct {
	Services []string       `json:"services,omitempty"`
	Log      logConf        `json:"log,omitempty"`
	Database databaseConfig `json:"database,omitempty"`
	Parsing  botParseConfig `json:"parsing,omitempty"`
}

type logConf struct {
	Level    string `json:"level,omitempty"`
	Location string `json:"location,omitempty"`
}

type databaseConfig struct {
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Pass     string `json:"pass,omitempty"`
	Database string `json:"database,omitempty"`
}

type botParseConfig struct {
	Reaction []string `json:"reaction,omitempty"`
	Response []string `json:"response,omitempty"`
	Max      int      `json:"max,omitempty"`
	AllowIP  bool     `json:"allow_ip,omitempty"`
}

func init() {
	flag.StringVarP(&verbose, "verbosity", "v", "info", "set the verbosity level for the bot {info,debug} (default is info)")
	flag.StringVarP(&logDir, "logdir", "l", "logs/", "set the log directory of the bot. (default is ./logs/)")
	flag.StringVarP(&confDir, "confdir", "d", "configs/", "set the config directory of the bot. (default is ./configs/)")
	flag.StringVarP(&conf, "conffile", "c", "parkertron.yml", "set the config name for the bot. (default is parkertron.yml)")
	flag.BoolVar(&diag, "diag", false, "uploads diagnotics to hastebin")
	flag.Parse()

	if !strings.HasSuffix(confDir, "/") {
		confDir = confDir + "/"
	}

	if newbot, err := loadInitConfig(confDir, conf, verbose); err != nil {
		log.Fatal(err)
	} else {
		if !flag.CommandLine.Changed(verbose) {
			verbose = newbot.Log.Level
		}

		if !flag.CommandLine.Changed(logDir) {
			logDir = newbot.Log.Location
		}
	}

	if diag {
		uploadDiag(logDir)
	}

	log.Print("starting logging")
	Log = newLogger(logDir, verbose)
	Log.Infof("logging online\n")

	if err := initConfig(confDir); err != nil {
		Log.Panic(err)
	}

	Log.Infof("%s %s\n\n", asciiArt, version)
}

func main() {
	// if there are no bots configured write default example configs
	if len(discordGlobal.Bots) == 0 && len(ircGlobal.Bots) == 0 {
		Log.Infof("No bots are configured")
		for _, service := range botConfig.Services {
			switch service {
			case "discord":
				if err := createExampleDiscordConfig(confDir + "discord/"); err != nil {
					Log.Fatalf("%s", err)
				}
			case "irc":
				if err := createExampleIRCConfig(confDir + "irc/"); err != nil {
					Log.Fatalf("%s", err)
				}
			default:
			}
		}
		Log.Infof("Example configs have been created.")
		Log.Info("shutting down")
		os.Exit(0)
	}

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

	Log.Infof("Bot is now running. Send 'shutdown' or 'ctrl + c' to stop the bot.\n")

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
	Log.Debugf("interupt caught\n")
	shutdown <- ""
}

func newLogger(logDir, level string) *logrus.Logger {
	if Log != nil {
		return Log
	}

	if _, err := os.Stat(logDir + "latest.log"); err != nil {
	} else {
		if err := os.Rename(logDir+"latest.log", logDir+time.Now().Format(time.RFC3339)+".log"); err != nil {
			Log.Errorf("there was an error opening the logs: %s", err)
		}
	}

	if _, err := os.Stat(logDir + "debug.log"); err != nil {
	} else {
		if err := os.Rename(logDir+"debug.log", logDir+"debug-"+time.Now().Format(time.RFC3339)+".log"); err != nil {
			Log.Errorf("there was an error opening the logs: %s", err)
		}
	}

	pathMap := lfshook.PathMap{
		logrus.InfoLevel:  logDir + "latest.log",
		logrus.DebugLevel: logDir + "debug.log",
		logrus.ErrorLevel: logDir + "latest.log",
		logrus.FatalLevel: logDir + "latest.log",
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

func uploadDiag(logDir string) {
	log.Printf("uploading logs to hastebin")
	if _, err := os.Stat(logDir + "latest.log"); err != nil {
	} else {
		uploadFile(logDir + "latest.log")
	}

	if _, err := os.Stat(logDir + "debug.log"); err != nil {
	} else {
		uploadFile(logDir + "debug.log")
	}

	os.Exit(0)
}

func uploadFile(name string) {
	hasteClient := haste.NewHaste("https://ptero.co")
	data, err := ioutil.ReadFile(name)
	if err != nil {
		Log.Infof("Unable to read file: %s\n", err.Error())
		os.Exit(2)
	}

	resp, err := hasteClient.UploadBytes(data)
	if err != nil {
		Log.Infof("Error uploading: %s\n", err.Error())
		os.Exit(3)
	}

	fmt.Println(name, resp.GetLink(hasteClient))
}
