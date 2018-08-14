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
	"gopkg.in/h2non/filetype.v1"
	"mvdan.cc/xurls"
)

// DataPackage hopefully pass info among the packages a bit easier.
type DataPackage struct {
	Service    string   `json:"service,omitempty"`
	Message    string   `json:"message,omitempty"`
	MessageID  string   `json:"message_id,omitempty"`
	AuthorID   string   `json:"author_id,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	BotID      string   `json:"bot_id,omitempty"`
	ChannelID  string   `json:"channel_id,omitempty"`
	Attached   []string `json:"attached,omitempty"`
	Perms      bool     `json:"perms,omitempty"`
	Group      string   `json:"group,omitempty"`
	GuildID    string   `json:"guild_id,omitempty"`
	DMChannel  string   `json:"dm_channel,omitempty"`
	DMMessage  string   `json:"dm_message,omitempty"`
}

func matchImage(input string) bool {
	ip := getParsingImageFiletypes()

	for _, ro := range ip {
		if strings.Contains(input, ro) == true {
			debug("Image found with a " + ro + " format")
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
				superdebug(rm[x][y])
				if strings.Contains(input, getParsingPasteString(rm[x][y])) {
					debug("Matched on: " + rm[x][y])
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
					debug("Matched on: " + rm[x][y])
					return rm[x][y]
				}
			}
		}
	}
	return ""
}

func parseBin(domain string, filename string) string {
	info("Reading from " + getParsingPasteString(domain+".url"))

	debug("Filename is: " + filename)

	formatted := ""

	getParsingPasteString(domain + ".format")

	re := regexp.MustCompile("&([a-z]*)&")
	rm := re.FindAllStringSubmatch(getParsingPasteString(domain+".format"), -1)

	for x, ro := range rm {
		for y := range ro {
			if y == 1 {
				formatted = formatted + getParsingPasteString(domain+"."+rm[x][y])
			}
		}
	}

	rawURL := formatted + filename

	debug("Raw text URL is " + rawURL)

	resp, err := http.Get(rawURL)
	if err != nil {
		fatal("", err)
	}

	body, err := ioutil.ReadAll(resp.Body)

	content := string(body)

	superdebug("Contents = \n" + content)

	return content
}

func parseImage(remoteURL string) string {
	info("Reading from " + remoteURL)

	remote, err := http.Get(remoteURL)
	if err != nil {
		fatal("", err)
	}

	defer remote.Body.Close()
	lastBin := strings.LastIndex(remoteURL, "/")
	fileName := remoteURL[lastBin+1:]

	debug("Filename is " + fileName)

	//open a file for writing
	file, err := os.Create("/tmp/" + fileName)
	if err != nil {
		fatal("", err)
	}
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, remote.Body)
	if err != nil {
		fatal("", err)
	}

	file.Close()
	debug("Image File Pulled and saved to /tmp/" + fileName)

	buf, _ := ioutil.ReadFile("/tmp/" + fileName)

	if filetype.IsImage(buf) {
		debug("File is an image")
	} else {
		debug("File is not an image")
		return ""
	}

	client := gosseract.NewClient()
	defer client.Close()

	client.SetImage("/tmp/" + fileName)
	text, err := client.Text()
	if err != nil {
		fatal("", err)
	}

	text = text[:len(text)-1]
	debug(text)
	debug("Image Parsed")

	return text
}

func parseKeyword(dpack DataPackage) {

	debug("Parsing inbound chat")

	pasteMatched, pasteDomain := matchPasteDomain(dpack.Message)

	debug("Matched domain: " + strconv.FormatBool(pasteMatched))

	//Catch domains and route to the proper controllers (image, binsite parsers)
	for _, url := range xurls.Relaxed.FindStringSubmatch(dpack.Message) {
		superdebug(url)
	}

	if matchImage(dpack.Message) == true {
		if matchImage(xurls.Relaxed.FindString(dpack.Message)) {
			dpack.Message = parseImage(xurls.Relaxed.FindString(dpack.Message))
		}
	} else if pasteMatched == true {
		if xurls.Relaxed.FindString(dpack.Message) != "" {
			debug("Sending: " + pasteDomain)
			debug("xurls matched: " + xurls.Relaxed.FindString(dpack.Message))
			debug("Guessing file name is: " + strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
			dpack.Message = parseBin(pasteDomain, strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
		}
	}

	//exact match search
	debug("Testing exact matches")
	for _, kr := range getKeywords() {
		superdebug("Testing on '" + strings.TrimPrefix(kr, "keyword.exact.") + "' and match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimPrefix(kr, "keyword.exact."))))
		if strings.ToLower(dpack.Message) == strings.TrimPrefix(kr, "keyword.exact.") == true {
			debug(getKeywordResponseString(kr))
			sendResponse(dpack, getKeywordResponseString(strings.TrimPrefix(kr, "keyword.")))
		}
	}

	lastKeyword := ""
	lastIndex := -1
	for _, kr := range getKeywords() {
		superdebug("Testing on '" + strings.TrimPrefix(kr, "keyword.") + "' and match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimPrefix(kr, "keyword."))))
		i := strings.LastIndex(strings.ToLower(dpack.Message), strings.TrimPrefix(kr, "keyword."))
		if i > lastIndex {
			lastIndex = i
			lastKeyword = kr
		}
	}
	if lastIndex > -1 {
		sendResponse(dpack, getKeywordResponseString(strings.TrimPrefix(lastKeyword, "keyword.")))
	}
	return
}

func parseAdminCommand(dpack DataPackage) {
	debug("Parsing inbound command: \n" + dpack.Message)
	if strings.HasPrefix(dpack.Message, "list") {
		debug("Getting available commands")
		req := strings.TrimPrefix(dpack.Message, "list ")
		response = "This is the list of current " + req + ": " + getCommandsString()
		if req == "commands" {
			sendResponse(dpack, "This is the list of current "+req+": "+getCommandsString())
			return
		} else if req == "keywords" {
			sendResponse(dpack, "This is the list of current "+req+": "+getKeywordsString())
			return
		} else {
			sendResponse(dpack, "There was no match for "+req+" options")
			return
		}
	}
}

func parseModCommand(dpack DataPackage) {
	debug("Parsing inbound command: \n" + dpack.Message)
}

func parseCommand(dpack DataPackage) {
	debug("Parsing inbound command: \n" + dpack.Message)

	if strings.HasPrefix(dpack.Message, "ggl") == true {
		debug("Googling for user. \n")
		sendResponse(dpack, "<https://lmgtfy.com/?q="+strings.Replace(strings.TrimPrefix(dpack.Message, "ggl "), " ", "+", -1)+">")
		return
	}

	//Search command file for command and prep response
	for _, cr := range getCommands() {
		if strings.Contains(strings.TrimPrefix(cr, "command."), dpack.Message) == true {
			sendResponse(dpack, getCommandResponseString(dpack.Message))
			return
		}
	}
}

func sendResponse(dpack DataPackage, response string) {
	if dpack.Service == "discord" {
		sendDiscordMessage(dpack, response)
	} else if dpack.Service == "irc" {
		sendIRCMessage(dpack.ChannelID, response)
	} else {
		return
	}
}
