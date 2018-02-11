package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/viper"
)

//Channel structure
type Channel struct {
	Chan string   `json:"channel"`
	CID  []string `json:"cid"`
}

func getChannels() []Channel {
	//Opens commands.json and returns values
	raw, err := ioutil.ReadFile("./listening.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var c []Channel
	json.Unmarshal(raw, &c)
	return c
}

//returns if user is blacklisted
func listening(a string) bool {
	channels := getChannels()

	for _, c := range channels {
		if c.Chan == "channel" {
			for _, u := range c.CID {
				if u == a {
					return true
				}
			}
		}
	}
	return false
}

func listenon(channel string) bool {
	if viper.GetBool("per_chan") == false {
		return true
	} else if listening(channel) == false {
		return false
	} else {
		return true
	}
}
