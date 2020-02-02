package main

import (
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/h2non/filetype"
	"github.com/otiai10/gosseract"
	"mvdan.cc/xurls"
)

// image handling
func matchImage(input string, imageTypes parsingImageConfig) bool {
	for _, ro := range imageTypes.Filetypes {
		if strings.Contains(input, ro) {
			Log.Debug("Image found with a " + ro + " format")
			return true
		}
	}
	return false
}

func parseImage(remoteURL string) string {
	Log.Info("Reading from " + remoteURL)

	remote, err := http.Get(remoteURL)
	if err != nil {
		Log.Fatal("", err)
	}

	defer remote.Body.Close()
	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := remoteURL[lastBin+1:]

	Log.Debug("Filename is " + fileName)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		Log.Fatal("", err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		Log.Fatal("", err)
	}

	file.Close()
	Log.Debug("Image File Pulled and saved to /tmp/" + fileName)

	buf, _ := ioutil.ReadFile("/tmp/" + fileName)

	if filetype.IsImage(buf) {
		Log.Debug("File is an image")
	} else {
		Log.Debug("File is not an image")
		return ""
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("/tmp/" + fileName)
	w, h := getImageDimension("/tmp/" + fileName)
	Log.Debug("Image width is " + strconv.Itoa(h))
	Log.Debug("Image height is " + strconv.Itoa(w))
	text, err := client.Text()
	if err != nil {
		Log.Fatal("", err)
	}

	text = text[:len(text)-1]
	Log.Debug(text)
	Log.Debug("Image Parsed")

	return text
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		Log.Fatal("error sending message", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		Log.Fatal("error sending message", err)
	}
	return image.Width, image.Height
}

// paste site handling
func parseBin(pasteConfig parsingPasteConfig, filename string) string {
	Log.Info("Reading from " + pasteConfig.Name)

	Log.Debug("Filename is: " + filename)

	Log.Debug("format is " + pasteConfig.Format)

	rawURL := strings.Replace(strings.Replace(pasteConfig.Format, "&url&", pasteConfig.URL, 1), "&filename&", filename, 1)

	Log.Debug("Raw text URL is " + rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		Log.Fatal("", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	content := string(body)

	Log.Debug("Contents = \n" + content)

	return content
}

//     __                               __
//    / /_____ __ ___    _____  _______/ /
//   /  '_/ -_) // / |/|/ / _ \/ __/ _  /
//  /_/\_\\__/\_, /|__,__/\___/_/  \_,_/
//  	     /___/
func parseKeyword(message string, attached []string, channelKeywords []keyword, parseConf parsing) (response, reaction []string) {

	Log.Debug("Parsing inbound chat")

	//Catch domains and route to the proper controllers (image, binsite parsers)
	Log.Debug("Matching on links in text")
	for _, url := range xurls.Relaxed.FindStringSubmatch(message) {
		Log.Debug(url)
	}

	// if attached != nil {
	// 	Log.Debug("Matching on Attached links")
	// 	for x := range attached {
	// 		if matchImage(attached[x], parseConf.Image) {
	// 			if matchImage(xurls.Relaxed.FindString(attached[x]), parseConf.Image) {
	// 				message = parseImage(xurls.Relaxed.FindString(attached[x]))
	// 			}
	// 		}
	// 	}
	// }

	// if matchImage(message, parseConf.Image) {
	// 	if matchImage(xurls.Relaxed.FindString(message), parseConf.Image) {
	// 		message = parseImage(xurls.Relaxed.FindString(message))
	// 	}
	// } else if pasteMatched {
	// 	matchedURL := xurls.Relaxed.FindString(message)
	// 	if matchedURL != "" {
	// 		Log.Debug("Sending: " + pasteDomain)
	// 		Log.Debug("xurls matched: " + matchedURL)
	// 		// TODO: actually fix this
	// 		_, fileName := filepath.Split(matchedURL)
	// 		Log.Debug("Guessing file name is: " + fileName)
	// 		// message = parseBin(, fileName)
	// 	}
	// }

	//exact match search
	Log.Debug("Testing exact matches")
	for _, keyWord := range channelKeywords {
		if strings.ToLower(message) == keyWord.Keyword && keyWord.Exact { // if the match was an exact match
			Log.Debugf("Response is %v", keyWord.Response)
			Log.Debugf("Reaction is %v", keyWord.Reaction)
			return keyWord.Response, keyWord.Reaction
		} else if strings.Contains(strings.ToLower(message), keyWord.Keyword) { // if the match was just a match
			Log.Debugf("Response is %v", keyWord.Response)
			Log.Debugf("Reaction is %v", keyWord.Reaction)
			return keyWord.Response, keyWord.Reaction
		}
	}

	lastIndex := -1

	//Match on errors
	Log.Debug("Testing error matches")

	for _, keyWord := range channelKeywords {
		if strings.Contains(strings.ToLower(message), keyWord.Keyword) {
			Log.Debugf("match is %s", keyWord.Keyword)
		}

		index := strings.LastIndex(strings.ToLower(message), keyWord.Keyword)
		if index > lastIndex {
			lastIndex = index
			response = keyWord.Response
			reaction = keyWord.Reaction
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
func adminCommand(servCommands, servKeywords []response, message string) (response, reaction []string) {
	message = strings.ToLower(message)

	return
}

// ModCommand commands are hard coded for now
func modCommand(message string) (response, reaction []string) {
	message = strings.ToLower(message)
	return
}

// Command parses commands
func parseCommand(message string, channelCommands []command) (response, reaction []string) {
	message = strings.ToLower(message)

	return
}
