package slackhandler

import (
	"fmt"
	"github.com/nlopes/slack"
	"os"
)

func SendMessage(messageText string) error {
	token := os.Getenv("SLACK_TOKEN")
	api := slack.New(token)
	_, _, err := api.PostMessage("core-infrastructure", messageText, slack.PostMessageParameters{
		Username: "Botty McBotface",
		AsUser: true,
		UnfurlMedia: true,
		UnfurlLinks: true,
		EscapeText: false,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}


	return nil
}

func SetChannelTopic(topicString string) error {
	return nil
}
