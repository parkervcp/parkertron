package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
