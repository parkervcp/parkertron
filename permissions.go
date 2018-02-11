package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

//Perms structure
type Perms struct {
	Grp string   `json:"group"`
	UID []string `json:"uid"`
}

func getPerms() []Perms {
	//Opens commands.json and returns values
	raw, err := ioutil.ReadFile("./permissions.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var c []Perms
	json.Unmarshal(raw, &c)
	return c
}

//returns if user is blacklisted
func blacklisted(a string) bool {
	perms := getPerms()

	for _, p := range perms {
		if p.Grp == "blacklist" {
			for _, u := range p.UID {
				if u == a {
					return true
				}
			}
		}
	}
	return false
}

//returns if user is an moderator
func getMod(a string) bool {
	perms := getPerms()

	for _, p := range perms {
		if p.Grp == "moderator" {
			for _, u := range p.UID {
				if u == a {
					return true
				}
			}
		}
	}
	return false
}

//returns if user is an admin
func getAdmin(a string) bool {
	perms := getPerms()

	for _, p := range perms {
		if p.Grp == "admin" {
			for _, u := range p.UID {
				if u == a {
					return true
				}
			}
		}
	}
	return false
}
