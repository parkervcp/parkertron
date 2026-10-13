package main

import (
	"fmt"

	"github.com/nlopes/slack"
)

var (
	apiToken = ""
)

// this is for testing.
func startSlackConnection() {
	var (
		postAsUserName  string
		postAsUserID    string
		postToUserName  string
		postToUserID    string
		postToChannelID string
	)

	api := slack.New(apiToken)

	authTest, err := api.AuthTest()
	if err != nil {
		fmt.Printf("Error getting channels: %s\n", err)
		return
	}
	// Post as the authenticated user.
	postAsUserName = authTest.User
	postAsUserID = authTest.UserID

	// Posting to DM with self causes a conversation with slackbot.
	postToUserName = authTest.User
	postToUserID = authTest.UserID

	Log.Info(fmt.Sprintf("Posting as %s (%s) in DM with %s (%s), channel %s\n", postAsUserName, postAsUserID, postToUserName, postToUserID, postToChannelID))
}
