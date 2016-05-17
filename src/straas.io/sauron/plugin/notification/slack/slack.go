package slack

import (
	"fmt"

	"github.com/bluele/slack"

	"straas.io/base/logger"
	"straas.io/sauron"
	"straas.io/sauron/plugin/notification"
)

var (
	log = logger.Get()
)

// NewSlackSinker creates a slack sinker
func NewSlackSinker() notification.Sinker {
	return &slackSinker{}
}

type slackCfg struct {
	// Token is slack api token
	Token string `json:"token" yaml:"token"`
	// UserName is message display user name
	UserName string `json:"user_name" yaml:"user_name"`
	// Channel is the channel name
	Channel string `json:"channel" yaml:"channel"`
}

type slackSinker struct {
}

func (s *slackSinker) Sink(rawConfig interface{}, severity sauron.Severity,
	recovery bool, desc string) error {

	cfg := rawConfig.(*slackCfg)
	api := slack.New(cfg.Token)

	channel, err := api.FindChannelByName(cfg.Channel)
	if err != nil {
		log.Errorf("Fail to find slack channel: channel:%s, err:%v",
			cfg.Channel, err)
		return fmt.Errorf("Fail to find slack channel: channel:%s, err:%v",
			cfg.Channel, err)
	}

	// decide color
	color := "danger"
	title := "Incident report"
	if recovery {
		color = "good"
		title = "Resolve report"
	}
	options := &slack.ChatPostMessageOpt{
		AsUser:   false,
		Username: cfg.UserName,
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Color: color,
				Title: desc,
			},
		},
	}
	return api.ChatPostMessage(channel.Id, title, options)
}

func (s *slackSinker) ConfigFactory() interface{} {
	return &slackCfg{}
}
