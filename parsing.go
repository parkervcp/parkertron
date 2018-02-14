package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/otiai10/gosseract"
	"mvdan.cc/xurls"
)

// isValidUrl tests a string to determine if it is a url or not.
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}

func parseKeyword(input string) string {

	writeLog("debug", "Parsing inbound chat", nil)

	if strings.Contains(input, ".png") == true || strings.Contains(input, ".jpg") == true {
		remoteURL := xurls.Relaxed().FindString(input)
		if isValidURL(remoteURL) == false {
			return ""
		}
		writeLog("debug", "Contains link to image", nil)
		input = parseImage(remoteURL)
	}

	if strings.Contains(input, "astebin") == true {
		remoteURL := xurls.Relaxed().FindString(input)
		if isValidURL(remoteURL) == false {
			return ""
		}
		writeLog("debug", "Is a bin link", nil)
		input = parseBin(remoteURL)
	}

	//Search keywords file for keyword and prep response
	for _, kr := range getKeywords() {
		writeLog("debug", "Testing on '"+strings.TrimPrefix(kr, "keyword.")+"' and match is "+strconv.FormatBool(strings.Contains(strings.ToLower(strings.TrimPrefix(kr, "keyword.")), input)), nil)
		if strings.Contains(strings.ToLower(strings.TrimPrefix(kr, "keyword.")), input) == true {
			writeLog("debug", getKeywordResponseString(kr), nil)
			response = getKeywordResponseString(input)
			writeLog("debug", "response: "+response, nil)
		}
	}

	return response
}

func parseCommand(input string) string {

	writeLog("debug", "Parsing inbound command: "+input, nil)

	if strings.HasPrefix(input, "ggl") == true {
		writeLog("debug", "Googling for user. \n", nil)
		response = "<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(input, "ggl "), " ", "+", -1) + ">"

	} else if strings.HasPrefix(input, "list") {
		req := strings.TrimPrefix(input, "list ")
		response = "This is the list of current " + req + ": "
		if req == "commands" {
			response = response + getCommandsString()
		} else if req == "keywords" {
			response = response + getKeywordsString()
		} else {
			response = "There was no match for " + req + " options"
		}
	} else {
		//Search command file for command and prep response
		for _, cr := range getCommands() {
			if strings.Contains(strings.TrimPrefix(cr, "command."), input) == true {
				response = getCommandResponseString(input)
			}
		}
	}

	if response == "" {

		return ""
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
