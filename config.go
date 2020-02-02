package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	BotName  string // veriable based on folder name
}

func initConfig(confDir string) (err error) {
	if err = loadConfDirs(confDir); err != nil {
		return nil
	}

	// sort files to load cirrectly
	Log.Debugf("Sorting files")
	sort.SliceStable(files.Files, func(i, j int) bool {
		return files.Files[i].Location > files.Files[j].Location
	})

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

func loadConfDirs(confdir string) (err error) {
	cleanConfDir := path.Clean(confDir)

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
					loadConf(file)
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
				return
			}

			for bid, bot := range discordGlobal.Bots {
				if bot.BotName == conf.BotName {
					for sid, server := range discordGlobal.Bots[bid].Servers {
						// if the server exists drop it and re-append config
						if server.ServerID == tempServer.ServerID {
							discordGlobal.Bots[bid].Servers[sid] = discordGlobal.Bots[bid].Servers[len(discordGlobal.Bots[bid].Servers)-1]
							discordGlobal.Bots[bid].Servers = discordGlobal.Bots[bid].Servers[:len(discordGlobal.Bots[bid].Servers)-1]
							discordGlobal.Bots[bid].Servers = append(discordGlobal.Bots[bid].Servers, tempServer)
							return
						}
					}
					// if the server isn't in the list append it.
					discordGlobal.Bots[bid].Servers = append(discordGlobal.Bots[bid].Servers, tempServer)
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
		// if err = loadFromFile(conf.Location, &tempBot); err != nil {
		// 	return
		// }
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
		Log.Debug("loading yaml file")
		if err = readYamlFromFile(file, iface); err != nil {
			Log.Error(err)
			return
		}
		Log.Debugf("%+v", iface)
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
	if err := writeJSONToFile(bytes, file); err != nil {
		return err
	}

	return nil
}

// File management
func writeJSONToFile(jdata []byte, file string) error {
	Log.Printf("updating file %s", file)
	// create a file with a supplied name
	if jsonFile, err := os.Create(file); err != nil {
		return err
	} else if _, err = jsonFile.Write(jdata); err != nil {
		return err
	}

	return nil
}

func readJSONFromFile(file string, iface interface{}) error {
	Log.Debugf("reading from json file: %s", file)
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

func writeYamlToFile(ydata []byte, file string) error {
	Log.Printf("updating file %s", file)
	// create a file with a supplied name
	yamlFile, err := os.Create(file)
	if err != nil {
		return err
	}

	if _, err = yamlFile.Write(ydata); err != nil {
		return err
	}

	return nil
}

func readYamlFromFile(file string, iface interface{}) error {
	Log.Debugf("reading from yaml file: %s", file)
	// Log.Debugf("opening yaml file\n")
	yamlFile, err := os.Open(file)

	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}

	// Log.Debugf("holding file open\n")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		if err := yamlFile.Close(); err != nil {
			Log.Printf("Error while closing yaml file: %+v", err)
		}
	}()

	// Log.Printf("reading file\n")
	byteValue, _ := ioutil.ReadAll(yamlFile)
	if err = yaml.Unmarshal(byteValue, iface); err != nil {
		return err
	}

	return nil
}
