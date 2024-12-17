package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/h2non/filetype"
	"github.com/otiai10/gosseract/v2"
)

func parseImage(remoteURL string) (imageText string, err error) {
	Log.Info("Reading from " + remoteURL)

	remote, err := http.Get(remoteURL)
	if err != nil {
		return
	}

	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := strings.Split(remoteURL[lastBin+1:], "?")[0]

	if len(fileName) > 150 {
		fileName = fileName[len(fileName)-50:]
	}

	Log.Debug("Filename is " + fileName)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		return
	}

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		return
	}

	err = remote.Body.Close()
	if err != nil {
		return
	}

	err = file.Close()
	if err != nil {
		return
	}

	Log.Debug("Image File Pulled and saved to /tmp/" + fileName)

	//load file to read
	buf, err := os.ReadFile("/tmp/" + fileName)
	if err != nil {
		return
	}

	// check filetype
	if !filetype.IsImage(buf) {
		Log.Debugf("file is not an image\n")
		return
	}

	Log.Debug("File is an image")

	client := gosseract.NewClient()

	err = client.SetImage("/tmp/" + fileName)
	if err != nil {
		return
	}

	imageData, err := getImageDimension("/tmp/" + fileName)
	if err != nil {
		return
	}

	Log.Debugf("Image width is %d", imageData.Width)
	Log.Debugf("Image height is %d", imageData.Height)

	imageText, err = client.Text()
	if err != nil {
		return
	}

	if len(imageText) >= 1 {
		imageText = imageText[:len(imageText)-1]
	}

	err = client.Close()
	if err != nil {
		return
	}

	err = os.Remove("/tmp/" + fileName)

	Log.Debug("Image Parsed")
	Log.Debug(imageText)

	return
}

func getImageDimension(imagePath string) (imageData image.Config, err error) {
	file, err := os.Open(imagePath)
	if err != nil {
		Log.Error("error opening file")
		return
	}
	defer file.Close()

	imageData, _, err = image.DecodeConfig(file)
	if err != nil {
		Log.Error("error decoding image")
		return
	}

	return
}

// paste site handling
func parseBin(url, format string) (binText string, err error) {
	var rawURL string

	Log.Debugf("reading from %s", url)
	_, file := path.Split(url)
	rawURL = strings.Replace(format, "&filename&", file, 1)

	Log.Debug("Raw text URL is " + rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	binText = string(body)

	Log.Debug("Contents = \n" + binText)

	return
}

// parses url contents for images and paste sites.
func parseURL(url string, parseConf parsing) (parsedText string) {
	//Catch domains and route to the proper controllers (image, binsite parsers)
	Log.Debugf("checking for pastes and images on %s\n", url)
	// If the url ends with a / remove it. Stupid chrome adds them.
	if strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}

	//check for image filetypes
	Log.Debug("checking if image")
	for _, filetype := range parseConf.Image.FileTypes {
		// need to remove any flags set for the url by cutting anything from the ? to the end
		if strings.HasSuffix(strings.Split(url, "?")[0], filetype) {
			Log.Debug("found image file")
			if imageText, err := parseImage(url); err != nil {
				Log.Errorf("%s\n", err)
			} else {
				Log.Debugf(imageText)
				parsedText = imageText
				return
			}
		}
	}

	// check for paste sites
	Log.Debug("checking if bin file")
	for _, paste := range parseConf.Paste.Sites {
		if strings.HasPrefix(url, paste.URL) {
			if binText, err := parseBin(url, paste.Format); err != nil {
				Log.Errorf("%s\n", err)
			} else {
				Log.Debugf(binText)
				parsedText = binText
				return
			}
		}
	}

	return
}

//     __                               __
//    / /_____ __ ___    _____  _______/ /
//   /  '_/ -_) // / |/|/ / _ \/ __/ _  /
//  /_/\_\\__/\_, /|__,__/\___/_/  \_,_/
//  	     /___/

// returns response and reaction for keywords
func parseKeyword(message, botName string, channelKeywords []keyword, parseConf parsing) (response, reaction []string) {
	Log.Debugf("Parsing inbound chat for %s", botName)

	message = strings.ToLower(message)

	//exact match search
	Log.Debug("Testing matches")
	for _, keyWord := range channelKeywords {
		if message == keyWord.Keyword && keyWord.Exact { // if the match was an exact match
			Log.Debugf("Response is %v", keyWord.Response)
			Log.Debugf("Reaction is %v", keyWord.Reaction)
			return keyWord.Response, keyWord.Reaction
		} else if strings.Contains(message, keyWord.Keyword) && !keyWord.Exact { // if the match was just a match
			Log.Debugf("Response is %v", keyWord.Response)
			Log.Debugf("Reaction is %v", keyWord.Reaction)
			return keyWord.Response, keyWord.Reaction
		}
	}

	lastIndex := -1

	//Match on errors
	Log.Debug("Testing matches")

	for _, keyWord := range channelKeywords {
		if strings.Contains(message, keyWord.Keyword) {
			Log.Debugf("match is %s", keyWord.Keyword)
		}

		index := strings.LastIndex(message, keyWord.Keyword)
		if index > lastIndex && !keyWord.Exact {
			lastIndex = index
			response = keyWord.Response
			reaction = keyWord.Reaction
		}
	}

	return
}

// returns response and reaction for patterns
func parseRegex(message, botName string, channelPatterns []pattern, parseConf parsing) (response, reaction []string) {
	Log.Debugf("Parsing inbound chat for %s", botName)

	message = strings.ToLower(message)

	//regex match search
	Log.Debug("Testing regex patterns")

	for _, pat := range channelPatterns {
		Log.Debugf("Pattern is %s", pat.Pattern)
		if match, err := regexp.MatchString(pat.Pattern, message); err != nil {
			Log.Error(err)
		} else if match {
			// if the pattern was a match
			Log.Debugf("Response is %v", pat.Response)
			Log.Debugf("Reaction is %v", pat.Reaction)
			return pat.Response, pat.Reaction
		}
	}
	return
}

//                                     __
//  _______  __ _  __ _  ___ ____  ___/ /
// / __/ _ \/  ' \/  ' \/ _ `/ _ \/ _  /
// \__/\___/_/_/_/_/_/_/\_,_/_//_/\_,_/
//

// AdminCommand commands are hard coded for now
func adminCommand(message, botName string, servCommands []command, servKeywords []keyword) (response, reaction []string) {
	Log.Debugf("Parsing inbound admin command for %s", botName)
	message = strings.ToLower(message)

	return
}

// ModCommand commands are hard coded for now
func modCommand(message, botName string, servCommands []command) (response, reaction []string) {
	Log.Debugf("Parsing inbound mod command for %s", botName)
	message = strings.ToLower(message)
	return
}

// Command parses commands
func parseCommand(message, botName string, channelCommands []command) (response, reaction []string) {
	Log.Debugf("Parsing inbound command for %s", botName)
	message = strings.ToLower(message)

	for _, command := range channelCommands {
		if command.Command == message {
			response = command.Response
			reaction = command.Reaction
		}
	}
	return
}

// general funcs
func contains(array []string, str string) bool {
	for _, value := range array {
		if value == str {
			return true
		}
	}
	return false
}
