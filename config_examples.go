package main

import (
	"log"
	"os"
	"strings"
)

func createExampleBotConfig(confDir, conf, verbose string) (err error) {
	newBot := parkertron{
		Services: []string{"discord"},
		Log: logConf{
			Level:    "info",
			Location: "logs/",
		},
		Parsing: botParseConfig{
			Max:      5,
			Response: []string{"There were too many logs to read &user&. Please post 5 or less."},
		},
	}

	// create file
	fileCheck, err := os.OpenFile(confDir+conf, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	if err := fileCheck.Close(); err != nil {
		return err
	}

	file, err := os.Stat(confDir + conf)
	if err != nil {
		return
	}

	if file.Size() == 0 {
		if strings.HasSuffix(conf, "yaml") || strings.HasSuffix(conf, "yml") {
			// if config is yaml
			if verbose == "debug" {
				log.Printf("file %s%s is yaml", confDir, conf)
			}

			if verbose == "debug" {
				log.Printf("writing to %s%s", confDir, conf)
			}
			if err = writeYamlToFile(confDir+conf, newBot); err != nil {
				return
			}
		} else if strings.HasSuffix(conf, "json") {
			// if config is json
			if verbose == "debug" {
				log.Printf("file %s%s is json", confDir, conf)
			}

			if verbose == "debug" {
				log.Printf("writing to %s%s", confDir, conf)
			}

			if err = writeJSONToFile(confDir+conf, newBot); err != nil {
				return
			}
		}
	}

	return
}

func createExampleDiscordConfig(confDir string) (err error) {
	// if the config dir doesn't exist make it
	Log.Debugf("creating example config folder %s if it doesn't exist", confDir)
	if err = createIfDoesntExist(confDir); err != nil {
		return
	}

	// if the config dir doesn't exist make it
	Log.Debugf("creating example config folder %s if it doesn't exist", confDir+"example-bot/")
	if err = createIfDoesntExist(confDir + "example-bot/"); err != nil {

	}

	// if the config dir doesn't exist make it
	Log.Debugf("creating example config file %s if it doesn't exist", confDir+"example-bot/example.yml")
	if err = createIfDoesntExist(confDir + "example-bot/example.yml"); err != nil {
		return
	}

	newDiscordBot := discordBot{}

	newDiscordBotConfig := discordBotConfig{
		Token: "An example token",
		Game:  "Supporting Humans",
		DMResp: responseArray{
			Response: []string{""},
			Reaction: []string{""},
		},
	}

	Log.Debugf("writing example config to file %s", confDir+"example-bot/example.yml")

	newDiscordBot.Config = newDiscordBotConfig

	if err = writeYamlToFile(confDir+"example-bot/example.yml", newDiscordBot); err != nil {
		return
	}

	// if the config dir doesn't exist make it
	Log.Debugf("creating example server config folder %s if it doesn't exist", confDir+"example-bot/example-server/")
	if err = createIfDoesntExist(confDir + "example-bot/example-server/"); err != nil {
		return
	}

	newServer := discordServer{
		ServerID: "a-server-id",
		Config: discordServerConfig{
			Prefix: ".",
			Clear:  true,
		},
		ChanGroups: []channelGroup{
			{
				ChannelIDs: []string{
					"a-channel-id",
					"another-channel-id",
				},
				Mentions: mentions{
					Ping: responseArray{
						Reaction: []string{"👋"},
						Response: []string{"I see I was pinged.", "Please just post the issue you are having", " Or you can gather your logs and post them"},
					},
					Mention: responseArray{
						Reaction: []string{"👋"},
						Response: []string{""},
					},
				},
				Commands: []command{
					{
						Command:  "example",
						Response: []string{"This is the response to the &prefix&example command"},
					},
				},
				Keywords: []keyword{
					{
						Keyword:  "example",
						Reaction: []string{""},
						Response: []string{"I have responded to seeing the word example."},
					},
				},
				Regex: []pattern{
					{
						Pattern:  ".*example.*",
						Reaction: []string{""},
						Response: []string{"I have found the word example somewhere in there."},
					},
				},
				Parsing: parsing{
					Image: parsingImageConfig{
						FileTypes: []string{
							"png",
							"jpg"},
						Sites: []parsingConfig{
							{
								Name:   "pastebin",
								URL:    "'https://pastebin.com/'",
								Format: "'https://pastebin.com/raw/&filename&'",
							},
							{
								Name:   "hastebin",
								URL:    "'https://hastebin.com/'",
								Format: "'https://hastebin.com/raw/&filename&'",
							},
						},
					},
					Paste: parsingPasteConfig{
						Sites: []parsingConfig{},
					},
				},
				KOM: discordKickOnMention{},
			},
		},
		Permissions: []permission{
			{
				Group:       "admin",
				Users:       []string{},
				Roles:       []string{},
				Blacklisted: false,
			},
		},
		Filters: []filter{
			{
				Term: "a bad word",
				Reason: []string{
					"the message was removed because it had 'a bad word' in it",
				},
			},
		},
	}

	Log.Debugf("writing example server config to file %s", confDir+"example-bot/example-server/example.yml")
	if err = writeYamlToFile(confDir+"example-bot/example-server/example-server.yml", newServer); err != nil {
		return
	}

	return
}

func createExampleIRCConfig(confDir string) (err error) {
	// if the config dir doesn't exist make it
	Log.Debugf("creating example config folder %s if it doesn't exist", confDir)
	if err = createIfDoesntExist(confDir); err != nil {
		return
	}

	// if the config dir doesn't exist make it
	Log.Debugf("creating example config folder %s if it doesn't exist", confDir+"example-bot/")
	if err = createIfDoesntExist(confDir + "example-bot/"); err != nil {

	}

	// if the config dir doesn't exist make it
	Log.Debugf("creating example config file %s if it doesn't exist", confDir+"example-bot/example-bot.yml")
	if err = createIfDoesntExist(confDir + "example-bot/example.yml"); err != nil {
		return
	}

	newIrc := ircBot{
		Config: ircBotConfig{
			Server: ircServerConfig{
				Address:   "irc.freenode.net",
				Port:      6667,
				SSLEnable: true,
				Ident:     "parkertron",
				Password:  "Set-Your-Own",
				Nickname:  "parkertron",
				RealName:  "Parker McBot",
			},
			DMResp: responseArray{
				Response: []string{},
			},
			Prefix: ".",
		},
		ChanGroups: []channelGroup{
			{
				ChannelIDs: []string{
					"a-channel-name",
					"another-channel-name",
				},
				Mentions: mentions{
					Ping: responseArray{
						Response: []string{},
					},
					Mention: responseArray{
						Response: []string{},
					},
				},
				Commands: []command{
					{
						Command:  "example",
						Reaction: []string{""},
						Response: []string{"This is the response to the &prefix&example command"},
					},
				},
				Keywords: []keyword{
					{
						Keyword:  "example",
						Response: []string{"I have responded to seeing the word example."},
					},
				},
				Regex: []pattern{
					{
						Pattern:  ".*example.*",
						Reaction: []string{""},
						Response: []string{"I have found the word example somewhere in there."},
					},
				},
				Parsing: parsing{
					Image: parsingImageConfig{
						FileTypes: []string{},
						Sites:     []parsingConfig{},
					},
					Paste: parsingPasteConfig{
						Sites: []parsingConfig{},
					},
				},
				Permissions: []permission{
					{
						Group:       "admin",
						Users:       []string{},
						Roles:       []string{},
						Blacklisted: false,
					},
				},
			},
		},
	}

	Log.Debugf("writing example config to file %s", confDir+"example-bot/example-bot.yml")

	if err = writeYamlToFile(confDir+"example-bot/example.yml", newIrc); err != nil {
		return
	}

	return
}
