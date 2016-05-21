package gosak

import (
	"fmt"
	"log"

	"github.com/bluele/slack"
)

const (
	slackBotToken = "xoxb-18018736162-pOGrLi9GaD3Vm8tfooE74gnb"
)

// ListSlackChannels lists all slack channels
func ListSlackChannels() {
	api := slack.New(slackBotToken)
	channels, err := api.ChannelsList()
	if err != nil {
		log.Printf("Fail to list slack channels: err[%s]", err.Error())
	}
	for _, channel := range channels {
		fmt.Println(channel.Id, channel.Name)
	}
}

// PostSlackMessage posts message to some channel
func PostSlackMessage(channelName, message string) {
	api := slack.New(slackBotToken)

	channel, err := api.FindChannelByName(channelName)
	if err != nil {
		log.Printf("Fail to find slack channel: slackChannel[%s], err[%s]",
			channelName, err.Error())
	}

	options := &slack.ChatPostMessageOpt{
		AsUser:   true,
		Username: "notify-gril",
	}
	err = api.ChatPostMessage(channel.Id, message, options)
	if err != nil {
		log.Printf("Fail to post slack message: err[%s]", err.Error())
	}
}
