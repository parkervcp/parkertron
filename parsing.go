package main

import (
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/h2non/filetype"
	"github.com/otiai10/gosseract"
	"mvdan.cc/xurls"
)

// image handling
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

func matchImage(input string) bool {
	imageTpyes := getParsingImageFiletypes()

	for _, ro := range imageTpyes {
		if strings.Contains(input, ro) {
			Log.Debug("Image found with a " + ro + " format")
			return true
		}
	}
	return false
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
func parseBin(domain string, filename string) string {
	Log.Info("Reading from " + getParsingPasteString(domain+".url"))

	Log.Debug("Filename is: " + filename)

	urlformat := getParsingPasteString(domain + ".format")

	Log.Debug("format is " + urlformat)

	rawURL := strings.Replace(strings.Replace(strings.Replace(urlformat, "&url&", getParsingPasteString(domain+".URL"), 1), "&filename&", filename, 1), "&append&", getParsingPasteString(domain+".append"), 1)

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

func matchPasteDomain(input string) (bool, string) {
	// Watched for matched domains
	re := regexp.MustCompile("([a-z]*.(url))")
	rm := re.FindAllStringSubmatch(getParsingPasteKeys(), -1)

	for x, ro := range rm {
		for y := range ro {
			if y%2 == 0 && rm[x][y] != "url" {
				Log.Debug(rm[x][y])
				if strings.Contains(input, getParsingPasteString(rm[x][y])) {
					Log.Debug("Matched on: " + rm[x][y])
					return true, strings.Replace(rm[x][y], ".url", "", -1)
				}
			}
		}
	}
	return false, ""
}

func formatURL(input string) string {
	// Watched for matched domains
	re := regexp.MustCompile("&([a-z]*)&")
	rm := re.FindAllStringSubmatch(getParsingPasteKeys(), -1)

	for x, ro := range rm {
		for y := range ro {
			if y%2 == 0 && rm[x][y] != "url" {
				if strings.Contains(input, rm[x][y]) {
					Log.Debug("Matched on: " + rm[x][y])
					return rm[x][y]
				}
			}
		}
	}
	return ""
}

//     __                               __
//    / /_____ __ ___    _____  _______/ /
//   /  '_/ -_) // / |/|/ / _ \/ __/ _  /
//  /_/\_\\__/\_, /|__,__/\___/_/  \_,_/
//  	     /___/

func parseKeyword(dpack DataPackage) {

	Log.Debug("Parsing inbound chat")

	pasteMatched, pasteDomain := matchPasteDomain(dpack.Message)

	Log.Debug("Matched domain: " + strconv.FormatBool(pasteMatched))

	//Catch domains and route to the proper controllers (image, binsite parsers)
	Log.Debug("Matching on links in text")
	for _, url := range xurls.Relaxed.FindStringSubmatch(dpack.Message) {
		Log.Debug(url)
	}

	if dpack.Attached != nil {
		Log.Debug("Matching on Attached links")
		for x := range dpack.Attached {
			if matchImage(dpack.Attached[x]) {
				if matchImage(xurls.Relaxed.FindString(dpack.Attached[x])) {
					dpack.Message = parseImage(xurls.Relaxed.FindString(dpack.Attached[x]))
				}
			}
		}
	}

	if matchImage(dpack.Message) {
		if matchImage(xurls.Relaxed.FindString(dpack.Message)) {
			dpack.Message = parseImage(xurls.Relaxed.FindString(dpack.Message))
		}
	} else if pasteMatched {
		if xurls.Relaxed.FindString(dpack.Message) != "" {
			Log.Debug("Sending: " + pasteDomain)
			Log.Debug("xurls matched: " + xurls.Relaxed.FindString(dpack.Message))
			Log.Debug("Guessing file name is: " + strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
			dpack.Message = parseBin(pasteDomain, strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
		}
	}

	//exact match search
	Log.Debug("Testing exact matches")
	for _, kr := range getKeywords() {
		if strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response")) {
			Log.Debug(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response") + " match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response"))))
		}
		if strings.ToLower(dpack.Message) == strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response") {
			dpack.Response = getKeywordResponseString(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response"))
			dpack.Keyword = strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response")
			Log.Debug("Response is " + dpack.Response)
			sendResponse(dpack)
		}
	}

	lastKeyword := ""
	lastIndex := -1
	//Match on errors
	Log.Debug("Testing error matches")
	for _, kr := range getKeywords() {
		if strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response")) {
			Log.Debug(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response") + " match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response"))))
		}
		i := strings.LastIndex(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response"))
		if i > lastIndex {
			lastIndex = i
			lastKeyword = kr
			dpack.Keyword = strings.TrimSuffix(strings.TrimPrefix(lastKeyword, "keyword."), ".response")
		}
	}

	if lastIndex > -1 {
		dpack.Response = getKeywordResponseString(strings.TrimSuffix(strings.TrimPrefix(lastKeyword, "keyword."), ".response"))
		sendResponse(dpack)
	}
	return
}

//                                     __
//  _______  __ _  __ _  ___ ____  ___/ /
// / __/ _ \/  ' \/  ' \/ _ `/ _ \/ _  /
// \__/\___/_/_/_/_/_/_/\_,_/_//_/\_,_/
//

// AdminCommand commands are hard coded for now
func adminCommand(servCommands []match, servKeywords []match, message string) (response []string) {
	message = strings.ToLower(message)
	Log.Debugf("parsing inbound admin command: %s\n", message)
	if strings.HasPrefix(message, "list") {
		Log.Debugf("getting available %s\n", strings.TrimPrefix(message, "list "))
		req := strings.TrimPrefix(message, "list ")
		if req == "commands" {
			allCommands := []string{"All commands are as follows"}
			for _, command := range servCommands {
				allCommands = append(allCommands, command.Command)
			}
			return allCommands
		} else if req == "keywords" {
			allKeywords := []string{"All keywords are as follows"}
			for _, keyword := range servKeywords {
				for _, keywords := range keyword.Keywords {
					allKeywords = append(allKeywords, keywords)
				}
			}
			return allKeywords
		} else if req == "" {
			return []string{"I can only look up keywords and commands right now."}
		} else {
			return []string{"There was no match for " + req + " options "}
		}
	}
	return []string{}
}

// ModCommand commands are hard coded for now
func modCommand(servCommands []match, message string) (response []string) {
	message = strings.ToLower(message)
	Log.Debugf("parsing inbound mod command: %s", message)
	return []string{}
}

// Command parses commands
func parseCommand(servCommands match, message string) (response []string) {
	message = strings.ToLower(message)
	Log.Debugf("parsing inbound command: %s", message)

	if strings.HasPrefix(message, "ggl") {
		Log.Debugf("googling for user.\n")
		return []string{"<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(message, "ggl "), " ", "+", -1) + ">"}
	}

	for _, command := range servCommands.Match {
		if command == message {
			return command.Response
		}
	}

	return []string{}
}
