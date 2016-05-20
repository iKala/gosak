package pagerduty

import (
	"time"

	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/external"
	"straas.io/sauron"
	"straas.io/sauron/plugin/notification"
)

const (
	interval = 15 * time.Minute
	layout   = "2006-01-02_15:04"
)

var (
	log = logger.Get()
)

// NewSinker creates a pagerDuty sinker
func NewSinker(pd external.PagerDuty, clock timeutil.Clock) notification.Sinker {
	return &pagerDutySinker{
		api:   pd,
		clock: clock,
	}
}

type pagerDutyCfg struct {
	// UserName is message display user name
	ServiceKey string `json:"service_key" yaml:"service_key"`
}

type pagerDutySinker struct {
	api   external.PagerDuty
	clock timeutil.Clock
}

func (s *pagerDutySinker) Name() string {
	return "slack"
}

func (s *pagerDutySinker) Sink(rawConfig interface{}, severity sauron.Severity,
	recovery bool, desc string) error {
	// it's safe
	cfg := rawConfig.(*pagerDutyCfg)

	if recovery {
		log.Info("ignore recovery mode")
		// close ?! TBD
		return nil
	}

	// incase make too many incidents
	// create incident key by time
	incidentKey := makeIncidentKey(s.clock.Now())
	log.Infof("create pagerduty incident, service:%s, incident:%s, desc:%s",
		cfg.ServiceKey, incidentKey, desc)
	return s.api.Trigger(cfg.ServiceKey, incidentKey, desc)
}

func (s *pagerDutySinker) ConfigFactory() interface{} {
	return &pagerDutyCfg{}
}

func makeIncidentKey(ts time.Time) string {
	return "ik" + ts.Truncate(interval).Format(layout)
}
