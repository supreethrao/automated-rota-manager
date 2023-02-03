package slackhandler

import (
	"fmt"

	"github.com/nlopes/slack"
)

type Messager struct {
	privateSlackConfig
}

type SlackConfig struct {
	Token string
	Channel string
	UserName string
}

type privateSlackConfig struct {
	SlackConfig
}

func (m *Messager) SendMessage(messageText string) error {
	api := slack.New(m.Token)
	_, _, err := api.PostMessage(m.Channel, messageText, slack.PostMessageParameters{
		Username:    m.UserName,
		AsUser:      true,
		UnfurlMedia: true,
		UnfurlLinks: true,
		EscapeText:  false,
	})
	if err != nil {
		fmt.Println("Failed to post Slack message")
		fmt.Println(err)
		return err
	}

	return nil
}

func (m *Messager) SetChannelTopic(topicString string) error {
	api := slack.New(m.Token)
	_, err := api.SetChannelTopic(m.Channel, topicString)
	return err
}

func NewMessager(slackConfig SlackConfig) *Messager {
	return &Messager{
		privateSlackConfig{
			slackConfig,
		},
	}
}
