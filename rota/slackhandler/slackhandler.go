package slackhandler

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/sky-uk/support-bot/helpers"
)

func SendMessage(messageText string) error {
	token := helpers.Getenv("SLACK_TOKEN", "")
	channel := helpers.Getenv("SLACK_CHANNEL", "core-infrastructure")
	username := helpers.Getenv("SLACK_USERNAME", "Botty McBotface")
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
