package pagerduty

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/danryan/go-pagerduty/pagerduty"

	"straas.io/external"
)

const (
	triggerUrl = "https://events.pagerduty.com/generic/2010-04-15/create_event.json"
)

var (
	clients    = map[string]*pagerduty.Client{}
	clientLock = sync.Mutex{}
)

// New creates an pagerduty intance
func New() external.PagerDuty {
	return &impl{}
}

type PagerDuty interface {
	Trigger(
		token string,
		serviceKey string,
		incidentKey string,
		desc string,
	) error
}

type impl struct {
}

func (s *impl) Trigger(token string,
	serviceKey string,
	incidentKey string,
	desc string) error {
	client := ensureClient(token)

	data := map[string]interface{}{
		"service_key": serviceKey,
		"event_type":  "trigger",
		"description": desc,
	}
	if incidentKey != "" {
		data["incident_key"] = incidentKey
	}

	body := map[string]interface{}{}
	resp, err := client.Post(triggerUrl, data, &body)
	if err != nil {
		return fmt.Errorf("fail to trigger pagerduty incident, err:%v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to post pagerduty incident, status %d, resp:%+v",
			resp.StatusCode, body)
	}
	return nil
}

func ensureClient(token string) *pagerduty.Client {
	clientLock.Lock()
	defer clientLock.Unlock()

	client, ok := clients[token]
	if ok {
		return client
	}
	httpClient := &http.Client{
		Timeout: 15 * time.Second,
	}
	client = pagerduty.NewClient("events", token, httpClient)
	clients[token] = client
	return client
}
