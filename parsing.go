package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
)

func parseImage(remoteURL string) (imageText string, err error) {
	Log.Info("Reading from " + remoteURL)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", remoteURL, nil)
	if err != nil {
		return "", err
	}

	remote, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if remote.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error code %d", remote.StatusCode)
	}

	buf, err := io.ReadAll(remote.Body)

	// get information on the image
	imageData, _, err := image.DecodeConfig(remote.Body)
	if err != nil {
		return
	}

	Log.Debugf("Image width is %d", imageData.Width)
	Log.Debugf("Image height is %d", imageData.Height)

	client := gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(buf)
	if err != nil {
		return
	}

	imageText, err = client.Text()
	if err != nil {
		return
	}

	if len(imageText) > 0 {
		imageText = imageText[:len(imageText)-1]
	}

	Log.Debug("Image Parsed")
	Log.Debug(imageText)

	return
}

// paste site handling
func parseBin(url, format string) (binText string, err error) {
	var rawURL string

	Log.Debugf("reading from %s", url)
	_, file := path.Split(url)
	rawURL = strings.Replace(format, "&filename&", file, 1)

	Log.Debug("Raw URL text is " + rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		err = errors.New(resp.Status)
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading from %s", url)
	}

	binText = string(body)

	return
}

// parses url contents for images and paste sites.
func parseURL(url string, parseConf parsing) (parsedText string, err error) {
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
			if imageText, parseErr := parseImage(url); parseErr != nil {
				return "", fmt.Errorf("%s\n", parseErr)
			} else {
				Log.Debug(imageText)
				parsedText = imageText
				return
			}
		}
	}

	// check for paste sites
	Log.Debug("checking if bin file")
	for _, paste := range parseConf.Paste.Sites {
		if strings.HasPrefix(url, paste.URL) {
			if binText, parseErr := parseBin(url, paste.Format); parseErr != nil {
				Log.Errorf("%s\n", fmt.Errorf("%s\n", parseErr))
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
			return keyWord.Response, keyWord.Reaction
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

	for _, indvCommand := range channelCommands {
		if indvCommand.Command == message {
			return indvCommand.Response, indvCommand.Reaction
		}
	}
	return
}
