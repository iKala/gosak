package metric

import (
	"flag"

	"straas.io/base/metric"
	"straas.io/external"
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
	flag.StringVar(&s.tag, "common_metric_tag", "metric", "metric fluent tag tag")
}

func (s *service) New(get common.ServiceGetter) (interface{}, error) {
	fluent := get(common.Fluent).(external.Fluent)

	done := make(chan bool)
	metric.StartExport(fluent, s.tag, done)
	return done, nil
}

func (s *service) Dependencies() []common.ServiceType {
	return []common.ServiceType{common.Fluent}
}
