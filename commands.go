package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/otiai10/gosseract"
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
		writeLog("error", "error reading commands ", err)
		os.Exit(1)
	}
	var c []Commands
	json.Unmarshal(raw, &c)
	return c
}

func parseChat(input string) string {
	commands := getCommands()
	writeLog("debug", "Parsing inbound chat", nil)
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
	writeLog("debug", "Parsing inbound command", nil)
	//Search command file for command and prep response
	for _, p := range commands {
		if p.Cmd == strings.ToLower(strings.TrimPrefix(input, getBotConfigString("prefix"))) {
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
	writeLog("info", "Reading from "+remoteURL, nil)

	lastBin := strings.LastIndex(remoteURL, "/")

	binName := remoteURL[lastBin+1:]

	rawBin := strings.Trim(binName, ".")

	baseURL := strings.Replace(remoteURL, binName, "", -1)

	writeLog("debug", "Base URL is "+baseURL, nil)

	if baseURL == "" {
		writeLog("debug", "just the domain and no file", nil)
		return ""
	}

	rawURL := baseURL + "raw/" + rawBin

	writeLog("debug", "Raw text URL is "+rawURL, nil)

	resp, err := http.Get(rawURL)
	if err != nil {
		writeLog("fatal", "", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	content := string(body)

	writeLog("debug", "Contents = \n"+content, nil)

	return content
}

func parseImage(remoteURL string) string {
	writeLog("info", "Reading from "+remoteURL, nil)

	remote, e := http.Get(remoteURL)
	if e != nil {
		writeLog("fatal", "", e)
	}

	defer remote.Body.Close()
	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := remoteURL[lastBin+1:]

	writeLog("debug", "Filename is "+fileName, nil)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		writeLog("fatal", "", err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		writeLog("fatal", "", err)
	}

	file.Close()
	writeLog("debug", "Image File Pulled and saved to /tmp/"+fileName, nil)

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("/tmp/" + fileName)
	text, err := client.Text()
	if err != nil {
		writeLog("fatal", "", err)
	}

	text = text[:len(text)-1]
	writeLog("debug", text, nil)
	writeLog("debug", "Image Parsed", nil)

	return text
}
