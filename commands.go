package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/gosseract"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

var (
	nextSend = time.Now()
)

//Commands structure
type Commands struct {
	Cmd string   `json:"command"`
	Typ string   `json:"type"`
	Lns []string `json:"lines"`
}

func getCommands() []Commands {
	//Opens commands.json and returns values
	raw, err := ioutil.ReadFile("./commands.json")
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	var c []Commands
	json.Unmarshal(raw, &c)
	return c
}

func hasPrefix(a string) bool {
	writeLog("debug", "prefix: "+strconv.FormatBool(strings.HasPrefix(a, viper.GetString("prefix"))), nil)
	return strings.HasPrefix(a, viper.GetString("prefix"))
}

func parseChat(input string) string {
	commands := getCommands()
	log.Debug("Parsing chat")
	//Search command file for command and prep response
	for _, p := range commands {
		if strings.Contains(strings.ToLower(input), p.Cmd) {
			if p.Typ == "listen" {
				for _, line := range p.Lns {
					response = response + "\n" + line
				}
			}
		}
	}
	return response
}

func parseCommand(input string) string {
	commands := getCommands()
	log.Debug("Parsing command")
	//Search command file for command and prep response
	for _, p := range commands {
		if p.Cmd == strings.ToLower(strings.TrimPrefix(input, viper.GetString("prefix"))) {
			if p.Typ == "chat" {
				for _, line := range p.Lns {
					response = response + "\n" + line
				}
			}
		}
	}
	return response
}

func parseBin(remoteURL string) string {
	log.Info("Reading from " + remoteURL)

	lastBin := strings.LastIndex(remoteURL, "/")

	binName := remoteURL[lastBin+1:]

	rawBin := strings.Trim(binName, ".")

	baseURL := strings.Replace(remoteURL, binName, "", -1)

	log.Debug("Base URL is " + baseURL)

	if baseURL == "" {
		log.Debug("just the domain and no file")
		return ""
	}

	rawURL := baseURL + "raw/" + rawBin

	log.Debug("Raw text URL is " + rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	content := string(body)

	log.Debug("Contents = \n" + content)

	return content
}

func parseImage(remoteURL string) string {
	log.Info("Reading from " + remoteURL)

	remote, e := http.Get(remoteURL)
	if e != nil {
		log.Fatal(e)
	}

	defer remote.Body.Close()
	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := remoteURL[lastBin+1:]

	log.Info("Filename is " + fileName)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		log.Fatal(err)
	}

	file.Close()
	log.Debug("Image File Pulled and saved to /tmp/" + fileName)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("/tmp/" + fileName)
	text, err := client.Text()
	if err != nil {
		log.Fatal(err.Error())
	}

	text = text[:len(text)-1]
	log.Debug(text)
	log.Debug("Image Parsed")

	return text
}
