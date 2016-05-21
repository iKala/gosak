package gosak

import (
	"fmt"
	"log"

	"github.com/bluele/slack"
)

// ListSlackChannels lists all slack channels
func ListSlackChannels(token string) {
	api := slack.New(token)
	channels, err := api.ChannelsList()
	if err != nil {
		log.Printf("Fail to list slack channels: err[%s]", err.Error())
	}
	for _, channel := range channels {
		fmt.Println(channel.Id, channel.Name)
	}
}

// PostSlackMessage posts message to some channel
func PostSlackMessage(token, channelName, message string) {
	api := slack.New(token)

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
