package slackhandler

import (
	"fmt"
	"github.com/nlopes/slack"
	"os"
)

func SendMessage(messageText string) error {
	token := os.Getenv("SLACK_TOKEN")
	channel := getenv("SLACK_CHANNEL", "core-infrastructure")
	username := getenv("SLACK_USERNAME", "Botty McBotface")
	api := slack.New(token)
	_, _, err := api.PostMessage(channel, messageText, slack.PostMessageParameters{
		Username: username,
		AsUser: true,
		UnfurlMedia: true,
		UnfurlLinks: true,
		EscapeText: false,
	})
	if err != nil {
	    	fmt.Println("Failed to post Slack message")
		fmt.Println(err)
		return err
	}


	return nil
}

func SetChannelTopic(topicString string) error {
	return nil
}

func getenv(key, fallback string) string {
    value := os.Getenv(key)
    if len(value) == 0 {
        return fallback
    }
    return value
}
