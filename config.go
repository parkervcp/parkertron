package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/goccy/go-yaml"
)

// File management
func writeJSONToFile(jdata []byte, file string) error {
	Log.Printf("updating file %s", file)
	// create a file with a supplied name
	jsonFile, err := os.Create(file)
	if err != nil {
		return err
	}
	_, err = jsonFile.Write(jdata)
	if err != nil {
		return err
	}

	return nil
}
func readJSONFromFile(file string, v interface{}) error {

	if !doesExist(file) {
		Log.Printf("%s does not exist creating it", file)
		jsonFile, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("there was an error loading the file")
		}
		_, err = jsonFile.Write([]byte("{}"))
		if err != nil {
			return err
		}
	}

	// Open our jsonFile
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
			Log.Printf("Error while closing JSON file: %v", err)
		}
	}()

	// Log.Printf("reading file\n")
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
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
	_, err = yamlFile.Write(ydata)
	if err != nil {
		return err
	}

	return nil
}
func readYamlFromFile(file string, v interface{}) error {
	if !doesExist(file) {
		Log.Printf("%s does not exist creating it", file)
		yamlFile, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("there was an error loading the file")
		}
		_, err = yamlFile.Write([]byte(""))
		if err != nil {
			return err
		}
	}

	// Open our jsonFile
	// Log.Printf("opening json file\n")
	yamlFile, err := os.Open(file)

	// if we os.Open returns an error then handle it
	if err != nil {
		return err
	}

	// Log.Printf("holding file open\n")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func() {
		if err := yamlFile.Close(); err != nil {
			Log.Printf("Error while closing JSON file: %v", err)
		}
	}()

	// Log.Printf("reading file\n")
	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(yamlFile)
	err = yaml.Unmarshal(byteValue, &v)
	if err != nil {
		return err
	}

	// return the json byte value.
	return nil
}

func doesExist(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}
	return false
}

func loadConfigs(confDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		Log.Error("%s", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					fmt.Println("modified file:", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
			}
		}
	}()

	err = filepath.Walk(confDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			Log.Infof("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.IsDir() && strings.Contains(info.Name(), "example") {
			return filepath.SkipDir
		}

		if strings.HasPrefix(info.Name(), ".") || strings.Contains(info.Name(), "example") {
			return nil
		}

		if !info.IsDir() {
			fullPath, err := filepath.Abs(path)
			if err != nil {
				Log.Errorf("error converting path %s\n", err)
			}

			Log.Infof("loading config into memory")

			Log.Infof("Loading file '%s' for file watcher\n", fullPath)

			err = watcher.Add(fullPath)
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				switch event.Op {
				case fsnotify.Write:

				}
			}
		}
	}()

	Log.Info("watcher started\n")

	<-done
	return nil
}

// LoadConfig loads configs from a file to an interface
func loadFile(file string, v interface{}) error {
	if strings.HasSuffix(file, ".json") {
		readJSONFromFile(file, v)
	} else if strings.HasSuffix(file, ".yml") || strings.HasSuffix(file, ".yaml") {
		readYamlFromFile(file, v)
	} else {
		return errors.New("no supported file type located")
	}

	return nil
}

// SaveConfig saves interfaces to a file
func SaveConfig(file string, v interface{}) error {
	// Log.Printf("converting struct data to bytesfor %s", file)
	bytes, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return fmt.Errorf("there was an error converting the user data to json")
	}

	// Log.Printf("writing bytes to file")
	if err := writeJSONToFile(bytes, file); err != nil {
		return err
	}

	return nil
}
