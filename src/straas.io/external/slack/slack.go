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
	channelId, err := getChannelId(api, channelName)
	if err != nil {
		return err
	}

	// decide color
	options := &slack.ChatPostMessageOpt{
		AsUser:   true,
		Username: userName,
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Color: color,
				Title: message,
			},
		},
	}
	return api.ChatPostMessage(channelId, title, options)
}

func getChannelId(api *slack.Slack, name string) (string, error) {
	channel, err := api.FindChannelByName(name)
	if err == nil {
		return channel.Id, nil
	}
	group, err := api.FindGroupByName(name)
	if err != nil {
		return "", fmt.Errorf("Fail to find slack channel: channel:%s, err:%v",
			name, err)
	}
	return group.Id, nil
}
