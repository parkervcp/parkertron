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

	"github.com/otiai10/gosseract"
	"gopkg.in/h2non/filetype.v1"
	"mvdan.cc/xurls"
)

// DataPackage hopefully pass info among the packages a bit easier.
type DataPackage struct {
	Service     string   `json:"service,omitempty"`
	Message     string   `json:"message,omitempty"`
	MessageID   string   `json:"message_id,omitempty"`
	AuthorID    string   `json:"author_id,omitempty"`
	AuthorName  string   `json:"author_name,omitempty"`
	AuthorRoles []string `json:"author_roles,omitempty"`
	BotID       string   `json:"bot_id,omitempty"`
	ChannelID   string   `json:"channel_id,omitempty"`
	Attached    []string `json:"attached,omitempty"`
	Perms       bool     `json:"perms,omitempty"`
	Group       string   `json:"group,omitempty"`
	GuildID     string   `json:"guild_id,omitempty"`
	DMChannel   string   `json:"dm_channel,omitempty"`
	DMMessage   string   `json:"dm_message,omitempty"`
	MsgTye      string   `json:"msg_type,omitempty"`
	Matched     string   `json:"matched,omitempty"`
	Response    string   `json:"response,omitempty"`
	Mention     string   `json:"mention,omitempty"`
	Reaction    []string `json:"reaction,omitempty"`
	ReactAdd    bool     `json:"react_job,omitempty"`
	Keyword     string   `json:"keyword,omitempty"`
	Command     string   `json:"command,omitempty"`
}

func matchImage(input string) bool {
	ip := getParsingImageFiletypes()

	for _, ro := range ip {
		if strings.Contains(input, ro) {
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
	w, h := getImageDimension("/tmp/" + fileName)
	debug("Image width is " + strconv.Itoa(h))
	debug("Image height is " + strconv.Itoa(w))
	text, err := client.Text()
	if err != nil {
		fatal("", err)
	}

	text = text[:len(text)-1]
	debug(text)
	debug("Image Parsed")

	return text
}

func getImageDimension(imagePath string) (int, int) {
	file, err := os.Open(imagePath)
	if err != nil {
		fatal("error sending message", err)
	}

	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fatal("error sending message", err)
	}
	return image.Width, image.Height
}

func parseKeyword(dpack DataPackage) {

	debug("Parsing inbound chat")

	pasteMatched, pasteDomain := matchPasteDomain(dpack.Message)

	debug("Matched domain: " + strconv.FormatBool(pasteMatched))

	//Catch domains and route to the proper controllers (image, binsite parsers)
	superdebug("Matching on links in text")
	for _, url := range xurls.Relaxed.FindStringSubmatch(dpack.Message) {
		superdebug(url)
	}

	if dpack.Attached != nil {
		superdebug("Matching on Attached links")
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
			debug("Sending: " + pasteDomain)
			debug("xurls matched: " + xurls.Relaxed.FindString(dpack.Message))
			debug("Guessing file name is: " + strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
			dpack.Message = parseBin(pasteDomain, strings.Replace(xurls.Relaxed.FindString(dpack.Message), getParsingPasteString(pasteDomain+".url"), "", -1))
		}
	}

	//exact match search
	debug("Testing exact matches")
	for _, kr := range getKeywords() {
		if strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response")) {
			superdebug(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response") + " match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response"))))
		}
		if strings.ToLower(dpack.Message) == strings.TrimSuffix(strings.TrimPrefix(kr, "keyword.exact."), ".response") {
			dpack.Response = getKeywordResponseString(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response"))
			dpack.Keyword = strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response")
			superdebug("Response is " + dpack.Response)
			sendResponse(dpack)
		}
	}

	lastKeyword := ""
	lastIndex := -1
	//Match on errors
	debug("Testing error matches")
	for _, kr := range getKeywords() {
		if strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response")) {
			superdebug(strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response") + " match is " + strconv.FormatBool(strings.Contains(strings.ToLower(dpack.Message), strings.TrimSuffix(strings.TrimPrefix(kr, "keyword."), ".response"))))
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

// admin commands are hard coded for now
func parseAdminCommand(dpack DataPackage) {
	debug("Parsing inbound admin command: " + dpack.Message)
	if strings.HasPrefix(dpack.Message, "list") {
		debug("Getting available commands")
		req := strings.TrimPrefix(dpack.Message, "list ")
		response = "This is the list of current " + req + ": " + getCommandsString()
		if req == "commands" {
			dpack.Response = "This is the list of current " + req + ": " + getCommandsString()
			sendResponse(dpack)
			return
		} else if req == "keywords" {
			dpack.Response = "This is the list of current " + req + ": " + getKeywordsString()
			sendResponse(dpack)
			return
		} else {
			dpack.Response = "There was no match for " + req + " options"
			sendResponse(dpack)
			return
		}
	}
}

// mod commands are hard coded for now
func parseModCommand(dpack DataPackage) {
	debug("Parsing inbound mod command: " + dpack.Message)
}

func parseCommand(dpack DataPackage) {
	debug("Parsing inbound command: " + dpack.Message)

	//Let Me Google That For You parsing
	if strings.HasPrefix(dpack.Message, "ggl") {
		debug("Googling for user. \n")
		dpack.Response = "<https://lmgtfy.com/?q=" + strings.Replace(strings.TrimPrefix(dpack.Message, "ggl "), " ", "+", -1) + ">"
		sendResponse(dpack)
		return
	}

	//Search command file for command and prep response
	for _, cr := range getCommands() {
		superdebug("Testing for " + cr)
		if strings.Contains(strings.TrimPrefix(cr, "command."), dpack.Message) {
			debug("match on " + cr)
			dpack.Response = getCommandResponseString(dpack.Message)
			sendResponse(dpack)
			return
		}
	}
}

func sendResponse(dpack DataPackage) {
	if dpack.Service == "discord" {
		sendDiscordMessage(dpack)
	} else if dpack.Service == "irc" {
		sendIRCMessage(dpack.ChannelID, dpack.Response)
	} else {
		return
	}
}
