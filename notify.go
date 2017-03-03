package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// slack url
const slackApiUrl string = "https://slack.com/api/chat.postMessage"

func notify(message string) (err error) {
	// get from env
	slackToken := os.Getenv("SLACK_TOKEN")
	slackChannel := os.Getenv("SLACK_CHANNEL")

	if slackToken == "" || slackChannel == "" {
		err = nil
		return
	}

	// make post data
	slackPostData := url.Values{}
	slackPostData.Set("token", slackToken)
	slackPostData.Set("channel", slackChannel)
	slackPostData.Set("username", "SSL Deadline Checker")
	slackPostData.Set("text", message)
	slackPostData.Set("icon_emoji", ":squirrel:")

	// post slack
	client := &http.Client{}
	r, err := http.NewRequest("POST", fmt.Sprintf("%s", slackApiUrl), bytes.NewBufferString(slackPostData.Encode()))
	if err != nil {
		return
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(r)
	return
}
