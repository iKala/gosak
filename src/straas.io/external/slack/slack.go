package slack

import (
	"fmt"

	"github.com/bluele/slack"

	"straas.io/external"
)

// New creates a slack implementation
func New(token string) external.Slack {
	return &slackImpl{
		token: token,
	}
}

type slackImpl struct {
	token string
}

func (s *slackImpl) Post(channelName, userName,
	title, message, color string) error {
	api := slack.New(s.token)
	channel, err := api.FindChannelByName(channelName)
	if err != nil {
		return fmt.Errorf("Fail to find slack channel: channel:%s, err:%v",
			channelName, err)
	}

	// decide color
	options := &slack.ChatPostMessageOpt{
		AsUser:   false,
		Username: userName,
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Color: color,
				Title: message,
			},
		},
	}
	return api.ChatPostMessage(channel.Id, title, options)
}
