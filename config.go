package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
)

var (
	// stores info on files for easier loading
	files confFiles

	// literally to make sure files load in order
	fileLoad = make(chan string)
)

type confFiles struct {
	Files []confFile
}

type confFile struct {
	Location string // full file path can be a dir or file
	Context  string // conf, botConf, serverConf
	Service  string // parkertron, discord, irc, slack
	BotName  string // variable based on folder name
}

func initConfig(confDir string) (err error) {
	if err = loadConfDirs(confDir); err != nil {
		return nil
	}

	Log.Debug(files)

	// sort all files to load in the correct order.
	files = fileSort()

	Log.Debug(files)

	// for all files pass it to fsnotify.
	Log.Debugf("loading files into file watcher")
	for _, file := range files.Files {
		go loadNWatch(file)
		<-fileLoad
	}

	Log.Debugf("Files that were loaded.")
	for _, v := range files.Files {
		Log.Debugf("%+v", v)
	}

	return nil
}

func fileSort() (newFiles confFiles) {
	// sort bot config
	for _, file := range files.Files {
		if file.Context == "conf" {
			newFiles.Files = append(newFiles.Files, file)
		}
	}

	// sort server config
	for _, file := range files.Files {
		if file.Context == "botConf" {
			newFiles.Files = append(newFiles.Files, file)
		}
	}

	// sort channel config
	for _, file := range files.Files {
		if file.Context == "serverConf" {
			newFiles.Files = append(newFiles.Files, file)
		}
	}
	return
}

func loadConfDirs(confdir string) (err error) {
	cleanConfDir := path.Clean(confdir)

	// Log.Debugf("reading from %s", cleanConfDir)
	confFullPath, err := filepath.Abs(cleanConfDir)
	if err != nil {
		Log.Errorf("error converting path %s\n", err)
	}

	// walk config dir supplied in startup single dir deep.
	err = filepath.Walk(confDir, func(fpath string, info os.FileInfo, err error) error {
		// if there are errors log the error
		if err != nil {
			Log.Infof("prevent panic by handling failure accessing a path %q: %+v\n", fpath, err)
			return err
		}

		// Log.Debugf("walking '%s'", fpath)

		// if an object has example in the name skip it
		if strings.Contains(info.Name(), "example") || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		fileFullPath, err := filepath.Abs(fpath)
		if err != nil {
			Log.Errorf("error converting path %s\n", err)
		}

		// Log.Debugf("passing '%s' to depthCounter", fpath)
		depth, err := depthCounter(confFullPath, fileFullPath)
		if err != nil {
			return err
		}

		// don't add to the file struct if it's a folder
		if info.IsDir() {
			return nil
		}

		// Log.Debugf(fpath)
		if len(strings.Split(fpath, "/")) >= 4 {
			// Log.Debug(strings.Split(fpath, "/")[2])
			files.Files = append(files.Files, confFile{fileFullPath, getFileType(depth), getFileService(fileFullPath), strings.Split(fpath, "/")[2]})
		} else {
			files.Files = append(files.Files, confFile{fileFullPath, getFileType(depth), getFileService(fileFullPath), ""})
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func depthCounter(confFullPath, fileFullPath string) (depth int, err error) {
	// Log.Debugf("checking on file %s\n", fileFullPath)
	// get depth of folder/file
	for depth := 1; depth <= 4; depth++ {
		// Log.Debugf("checking on %d", depth)
		// each /* is another file depth.
		match, err := path.Match(confFullPath+strings.Repeat("/*", depth), fileFullPath)
		if err != nil {
			return 0, err
		}
		if match {
			// Log.Debugf("Match on depth %d", depth)
			return depth, nil
		}
	}

	return 0, nil
}

func getFileType(depth int) (fileType string) {
	switch depth {
	case 1:
		fileType = "conf"
	case 3:
		fileType = "botConf"
	case 4:
		fileType = "serverConf"
	}

	return fileType
}

func getFileService(filePath string) (service string) {
	if strings.Contains(filePath, "discord") {
		service = "discord"
	} else if strings.Contains(filePath, "irc") {
		service = "irc"
	} else if strings.Contains(filePath, "slack") {
		service = "slack"
	} else {
		service = "parkertron"
	}

	return service
}

func loadNWatch(file confFile) {
	Log.Debugf("Loading file")
	if err := loadConf(file); err != nil {
		Log.Error(err)
	}

	Log.Debugf("Setting up watcher")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Log.Errorf("%+v", err)
		return
	}

	Log.Debugf("defer closing watcher")
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Write:
					Log.Infof("file changed: %s", event.Name)
					if err := loadConf(file); err != nil {
						Log.Errorf("%+v", err)
					}
				}
			case err := <-watcher.Errors:
				Log.Errorf("%+v", err)
			}
		}
	}()

	if err := watcher.Add(file.Location); err != nil {
		Log.Error(err)
	}

	fileLoad <- ""
	<-done
}

func loadConf(conf confFile) (err error) {
	switch conf.Service {
	case "parkertron":
		if err = loadFromFile(conf.Location, &botConfig); err != nil {
			Log.Error()
		}

	// discord conf loading
	case "discord":
		if conf.Context == "botConf" {
			Log.Debugf("loading config for discord bot %s", conf.BotName)
			// set up new temp var for the bot
			var tempBot discordBot
			tempBot.BotName = conf.BotName

			if err = loadFromFile(conf.Location, &tempBot); err != nil {
				Log.Errorf("there was an error loading configs")
				Log.Error(err)
				return
			}
			for bid, bot := range discordGlobal.Bots {
				if bot.BotName == conf.BotName {
					discordGlobal.Bots[bid].Config.Game = tempBot.Config.Game
					discordGlobal.Bots[bid].Config.DMResp = tempBot.Config.DMResp
					return
				}
			}

			// add bot to discord bots array.
			discordGlobal.Bots = append(discordGlobal.Bots, tempBot)

		} else if conf.Context == "serverConf" {
			Log.Debugf("loading discord server config for %s", conf.BotName)
			// set up new temp var for the server
			var tempServer discordServer

			if err = loadFromFile(conf.Location, &tempServer); err != nil {
				Log.Error(err)
				return
			}

			for i := range tempServer.ChanGroups {
				Log.Infof("channel groups %+v", tempServer.ChanGroups[i].ChannelIDs)
			}

			for bid, bot := range discordGlobal.Bots {
				if bot.BotName == conf.BotName {
					for sid, server := range discordGlobal.Bots[bid].Servers {
						// if the server exists drop it and re-append config
						if server.ServerID == tempServer.ServerID {
							discordGlobal.Bots[bid].Servers[sid].ChanGroups = tempServer.ChanGroups
							discordGlobal.Bots[bid].Servers[sid].Config = tempServer.Config
							discordGlobal.Bots[bid].Servers[sid].Permissions = tempServer.Permissions
							return
						}
					}
					// if the server isn't in the list append it.
					discordGlobal.Bots[bid].Servers = append(discordGlobal.Bots[bid].Servers, tempServer)
					Log.Debugf("loaded server %s for bot %s", tempServer.ServerID, bot.BotName)
				}
			}
		}
	// irc conf loading
	case "irc":
		if conf.Context == "botConf" {
			Log.Debugf("loading config for irc bot %s", conf.BotName)

			var tempBot ircBot
			tempBot.BotName = conf.BotName

			// if there is an error loading the config return
			if err = loadFromFile(conf.Location, &tempBot); err != nil {
				return
			}

			// add bot to irc bots array
			ircGlobal.Bots = append(ircGlobal.Bots, tempBot)
		}
	// slack config loading
	case "slack":

	}

	return nil
}

// LoadConfig loads configs from a file to an interface
func loadFromFile(file string, iface interface{}) (err error) {
	if strings.HasSuffix(file, ".json") { // if json file
		Log.Debug("loading json file")
		if err = readJSONFromFile(file, iface); err != nil {
			Log.Error(err)
			return
		}
	} else if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") { // if yaml file
		Log.Debugf("loading yaml file %s", file)
		if err = readYamlFromFile(file, iface); err != nil {
			Log.Error(err)
			return
		}
		// Log.Debugf("interface %+v", iface)
	} else {
		return errors.New("no supported file type located")
	}

	return nil
}

// SaveConfig saves interfaces to a file
func saveConfig(file string, iface interface{}) error {
	// Log.Printf("converting struct data to bytesfor %s", file)
	bytes, err := json.MarshalIndent(iface, "", " ")
	if err != nil {
		return fmt.Errorf("there was an error converting the user data to json")
	}

	// Log.Printf("writing bytes to file")
	if err := writeJSONToFile(file, bytes); err != nil {
		return err
	}

	return nil
}

// File management
func writeJSONToFile(file string, iface interface{}) (err error) {
	jdata, err := json.MarshalIndent(iface, "", "  ")
	if err != nil {
		return
	}

	// create a file with a supplied name
	if jsonFile, err := os.Create(file); err != nil {
		return err
	} else if _, err = jsonFile.Write(jdata); err != nil {
		return err
	}

	return nil
}

func readJSONFromFile(file string, iface interface{}) error {
	// Log.Printf("opening json file\n")
	jsonFile, err := os.Open(file)
	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}

	// Log.Printf("holding file open\n")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		if err := jsonFile.Close(); err != nil {
			Log.Printf("Error while closing JSON file: %+v", err)
		}
	}()

	// Log.Printf("reading file\n")
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err = json.Unmarshal(byteValue, iface); err != nil {
		return err
	}

	// return the json byte value.
	return nil
}

func writeYamlToFile(file string, iface interface{}) (err error) {
	ydata, err := yaml.Marshal(iface)
	if err != nil {
		return
	}

	// create a file with a supplied name
	yamlFile, err := os.Create(file)
	if err != nil {
		return
	}

	if _, err = yamlFile.Write(ydata); err != nil {
		return
	}

	return
}

func readYamlFromFile(file string, iface interface{}) (err error) {
	// Log.Debugf("opening yaml file\n")
	yamlFile, err := os.Open(file)
	if err != nil {
		return
	}

	// Log.Debugf("holding file open\n")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		if err := yamlFile.Close(); err != nil {
			return
		}
	}()

	// Log.Printf("reading file\n")
	byteValue, _ := ioutil.ReadAll(yamlFile)
	if err = yaml.Unmarshal(byteValue, iface); err != nil {
		return
	}

	return
}

// Exists reports whether the named file or directory exists.
func createIfDoesntExist(name string) (err error) {
	p, file := path.Split(name)

	// if confdir exists carry on
	if _, err := os.Stat(name); err != nil {
		// if file doesn't exist
		if os.IsNotExist(err) {
			// stat 
			if _, err = os.Stat(name); err != nil {
				if file == "" {
					if err = os.Mkdir(p, 0755); err != nil {
					}
				} else {
					if fileCheck, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
					} else {
						if err := fileCheck.Close(); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return
}

func loadInitConfig(confDir, conf, verbose string) (botConfig parkertron, err error) {
	// All of this is pre-Logrus init
	if verbose == "debug" {
		log.Printf("Checking for dir at %s", confDir)
	}

	// if the config dir doesn't exist make it
	if err := createIfDoesntExist(confDir); err != nil {
		return botConfig, err
		// if we can't make a confdir log a fatal error.
	}

	// if confdir exists carry on
	info, err := os.Stat(confDir)
	if err != nil {
		return botConfig, err
	}

	// if not a directory error out
	if !info.IsDir() {
		return botConfig, errors.New("given file not directory")
	}

	if verbose == "debug" {
		log.Printf("loading initial bot config at %s%s", confDir, conf)
	}

	// if config doesn't exist make it.
	if err = createIfDoesntExist(confDir + conf); err != nil {
		if verbose == "debug" {
			log.Printf("creating config %s", confDir+conf)
			if err := createExampleBotConfig(confDir, conf, verbose); err != nil {
				return parkertron{}, err
			}
		}
	}

	// if conf file exists carry on
	if verbose == "debug" {
		log.Printf("file %s%s exists", confDir, conf)
	}

	// if confdir exists carry on
	file, err := os.Stat(confDir + conf)
	if err != nil {
		return botConfig, err
	}

	if file.Size() == 0 {
		if err := createExampleBotConfig(confDir, conf, verbose); err != nil {
			return parkertron{}, err
		}
	}

	if strings.HasSuffix(conf, "yaml") || strings.HasSuffix(conf, "yml") {
		if err = readYamlFromFile(confDir+conf, &botConfig); err != nil {
			return
		}
	} else if strings.HasSuffix(conf, "json") {
		if err = readJSONFromFile(confDir+conf, &botConfig); err != nil {
			return botConfig, err
		}
	}

	return
}
