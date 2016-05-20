package pagerduty

import (
	"fmt"
	"net/http"
	"time"

	"github.com/danryan/go-pagerduty/pagerduty"

	"straas.io/external"
)

const (
	triggerURL = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"
)

// New creates an pagerduty intance
func New(token string) external.PagerDuty {
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	client := pagerduty.NewClient("events", token, httpClient)
	return &impl{
		client: client,
	}
}

type impl struct {
	client *pagerduty.Client
}

func (s *impl) Trigger(serviceKey, incidentKey, desc string) error {
	data := map[string]interface{}{
		"service_key": serviceKey,
		"event_type":  "trigger",
		"description": desc,
	}
	if incidentKey != "" {
		data["incident_key"] = incidentKey
	}

	body := map[string]interface{}{}
	resp, err := s.client.Post(triggerURL, data, &body)
	if err != nil {
		return fmt.Errorf("fail to trigger pagerduty incident, err:%v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to post pagerduty incident, status %d, resp:%+v",
			resp.StatusCode, body)
	}
	return nil
}
