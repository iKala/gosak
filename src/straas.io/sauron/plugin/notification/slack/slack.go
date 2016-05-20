package slack

import (
	"straas.io/base/logger"
	"straas.io/external"
	"straas.io/sauron"
	"straas.io/sauron/plugin/notification"
)

var (
	log = logger.Get()
)

// NewSinker creates a slack sinker
func NewSinker(slack external.Slack) notification.Sinker {
	return &slackSinker{
		api: slack,
	}
}

type slackCfg struct {
	// UserName is message display user name
	UserName string `json:"user_name" yaml:"user_name"`
	// Channel is the channel name
	Channel string `json:"channel" yaml:"channel"`
}

type slackSinker struct {
	api external.Slack
}

func (s *slackSinker) Name() string {
	return "slack"
}

func (s *slackSinker) Sink(rawConfig interface{}, severity sauron.Severity,
	recovery bool, desc string) error {
	// it's safe
	cfg := rawConfig.(*slackCfg)

	// decide color
	color := "danger"
	title := "Incident report"
	if recovery {
		color = "good"
		title = "Resolve report"
	}
	return s.api.Post(
		cfg.Channel,
		cfg.UserName,
		title,
		desc,
		color,
	)
}

func (s *slackSinker) ConfigFactory() interface{} {
	return &slackCfg{}
}
