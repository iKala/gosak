package external

import (
	"time"

	"github.com/coreos/etcd/client"
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

// Etcd defines an interface for etcd operation
type Etcd interface {
	// Watch watches key recursively
	Watch(etcdKey string, afterIndex uint64, resp chan<- *client.Response, done <-chan bool) *client.Error
	// GetAndWatch gets the key recursively and then watch the key
	GetAndWatch(etcdKey string, resp chan<- *client.Response, done <-chan bool)
	// Get returns the response recursively with the given key
	Get(etcdKey string, recursive bool) (*client.Response, error)
	// Set sets the value to etcd
	Set(etcdKey, value string) (*client.Response, error)
	// Set sets the value to etcd with TTL
	SetWithTTL(etcdKey, value string, ttl time.Duration) (*client.Response, error)
	// RefreshTTL refreshes ttl
	RefreshTTL(etcdKey string, ttl time.Duration) (*client.Response, error)
	// IsNotFound checks if err is not found error
	IsNotFound(err error) bool
}
