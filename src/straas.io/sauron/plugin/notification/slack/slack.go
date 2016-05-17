package slack

// import (
// 	"fmt"
// )

// type slackCfg struct {
// 	// Url is slack Incoming Webhooks
// 	Url string `json:"url" yaml:"url"`
// 	// UserName is message display user name
// 	UserName string `json:"user_name" yaml:"user_name"`
// }

// type slackNotifier struct {
// }

// func (s *slackNotifier) Sink(config interface{}, severity Severity,
// 	recovery bool, desc string) error {

// 	// must success,
// 	cfg := config.(*slackCfg)

// 	// leverage superagent to send config ?!
// 	fmt.Println(cfg)

// 	return nil
// }

// func (s *slackNotifier) ConfigFactory() interface{} {
// 	return &slackCfg{}
// }
