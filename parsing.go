package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/otiai10/gosseract"
	"mvdan.cc/xurls"
)

func matchImage(input string) bool {
	ip := getParsingImageFiletypes()

	for _, ro := range ip {
		if strings.Contains(input, ro) == true {
			writeLog("debug", "Image found with a "+ro+" format", nil)
			return true
		}
	}
	return false
}

func matchPasteDomain(input string) (bool, string) {
	// Watched for matched domains
	re := regexp.MustCompile("([a-z]*.(url))")
	rm := re.FindAllStringSubmatch(getParsingPasteKeys(), -1)

	for x, ro := range rm {
		for y := range ro {
			if y%2 == 0 && rm[x][y] != "url" {
				writeLog("debug", rm[x][y], nil)
				if strings.Contains(input, getParsingPasteString(rm[x][y])) {
					writeLog("debug", "Matched on: "+rm[x][y], nil)
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
					writeLog("debug", "Matched on: "+rm[x][y], nil)
					return rm[x][y]
				}
			}
		}
	}
	return ""
}

func parseKeyword(input string) string {

	writeLog("debug", "Parsing inbound chat", nil)

	pasteMatched, pasteDomain := matchPasteDomain(input)

	writeLog("debug", "Matched domain: "+strconv.FormatBool(pasteMatched), nil)

	if matchImage(input) == true {
		writeLog("debug", "xurls matched: "+xurls.Relaxed.FindString(input), nil)
		if matchImage(xurls.Relaxed.FindString(input)) == true {
			input = parseImage(xurls.Relaxed.FindString(input))
		}
	} else if pasteMatched == true {
		writeLog("debug", "Sending: "+pasteDomain, nil)
		writeLog("debug", "xurls matched: "+xurls.Relaxed.FindString(input), nil)
		writeLog("debug", "Guessing file name is: "+strings.Replace(xurls.Relaxed.FindString(input), getParsingPasteString(pasteDomain+".url"), "", -1), nil)
		parseBin(pasteDomain, strings.Replace(xurls.Relaxed.FindString(input), getParsingPasteString(pasteDomain+".url"), "", -1))
	}

	//Search keywords file for keyword and prep response
	for _, kr := range getKeywords() {
		writeLog("debug", "Testing on '"+strings.TrimPrefix(kr, "keyword.")+"' and match is "+strconv.FormatBool(strings.Contains(input, strings.TrimPrefix(kr, "keyword."))), nil)
		if strings.Contains(input, strings.TrimPrefix(kr, "keyword.")) == true {
			writeLog("debug", getKeywordResponseString(kr), nil)
			response = getKeywordResponseString(strings.TrimPrefix(kr, "keyword."))
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

func parseBin(domain string, filename string) string {
	writeLog("info", "Reading from "+getParsingPasteString(domain+".url"), nil)

	writeLog("debug", "Filename is: "+filename, nil)

	formatted := ""

	getParsingPasteString(domain + ".format")

	re := regexp.MustCompile("&([a-z]*)&")
	rm := re.FindAllStringSubmatch(getParsingPasteString(domain+".format"), -1)

	for x, ro := range rm {
		for y := range ro {
			if y == 1 {
				writeLog("debug", rm[x][y], nil)
				formatted = formatted + getParsingPasteString(domain+"."+rm[x][y])
			}
		}
	}

	rawURL := formatted + filename

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
