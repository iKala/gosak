package metric

import (
	"flag"

	"straas.io/base/metric"
	"straas.io/service/common"
)

func init() {
	common.Register(&service{})
}

type service struct {
	tag string
}

func (s *service) Type() common.ServiceType {
	return common.MetricExporter
}

func (s *service) AddFlags() {
	flag.StringVar(&s.tag, "common.metric_tag", "metric", "metric fluent tag")
}

func (s *service) New(get common.ServiceGetter) (interface{}, error) {
	done := make(chan bool)
	metric.StartExport(get.Fluent(), s.tag, done)
	return func() {
		close(done)
	}, nil
}

func (s *service) Dependencies() []common.ServiceType {
	return []common.ServiceType{common.Fluent}
}
