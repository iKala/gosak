package metric

import (
	"github.com/csigo/metric"
	"github.com/facebookgo/stats"
)

// New creates an instance of metric
func New(module string) stats.Client {
	return metric.NewClient(module, "")
}

// NewWithPrefix creates an instance of metric with name prefix
func NewWithPrefix(module, prefix string) stats.Client {
	return metric.NewClient(module, prefix)
}
