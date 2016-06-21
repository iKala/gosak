package external

import (
	"time"
)

// Slack defines an interface for post slack message
type Slack interface {
	// Post posts message to slack
	Post(channelName, userName, title, message, color string) error
}

// PagerDuty defines an interface for trigger incidents to pagerduty
type PagerDuty interface {
	// Trigger creates a pagerduty incident
	Trigger(serviceKey string, incidentKey string, desc string) error
}

// Elasticsearch defines an interface for query and post elasticsearch
type Elasticsearch interface {
	// Scalar queries es to scalar number
	Scalar(indices []string, queryStr, timeField string, strat, end time.Time,
		field, op string, wildcard bool) (*float64, error)
}

// Stackdriver deinfes an interface for query stackdriver
type Stackdriver interface {
	// List lists stackdriver metric by filter
	List(project, filter, op string, start, end time.Time) ([]Point, error)
}

// Fluent deinfes an interface to post data to fluent
type Fluent interface {
	// Post posts data to fluent
	Post(tag string, v interface{})
}

// Point is stackdriver timeseries point
type Point struct {
	Start time.Time
	End   time.Time
	Value float64
}
