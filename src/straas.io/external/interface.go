package external

import (
	"time"
)

// Slack defines an interface for post slack message
type Slack interface {
	// Post posts message to slack
	Post(token, channelName, userName, title, message, color string) error
}

// Elasticsearch defines an interface for query and post elasticsearch
// TODO: impl es
type Elasticsearch interface {
	// Scalar queries es to scalar number
	Scalar(indices []string, query string, strat, end time.Time,
		field, op string, wildcard bool) (*float64, error)
	// Post posts elasticsearch messages
	Post(index string, v ...interface{}) error
}
