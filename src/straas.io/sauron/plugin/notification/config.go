package notification

import (
	"encoding/json"
	"errors"
	"fmt"

	"straas.io/sauron"
)

// Config is the root of notification configuration
type Config struct {
	Groups []*Group `json:"groups" yaml:"groups"`
}

// Groups defines the notification group
type Group struct {
	// Name is group name
	Name string `json:"name" yaml:"name"`
	// Desc is group description
	Desc string `json:"desc" yaml:"desc"`
	// Notifications are a list of raw sinks
	RawSinkers []*RawMessage `json:"sinkers" yaml:"sinkers"`
}

// BaseSinkCfg defines common fields of notification
type BaseSinkCfg struct {
	// Type is notificdation type name
	Type string `json:"type" yaml:"type"`
	// Severity indicates given severity levels to use
	// this notification for alerts
	Severity []sauron.Severity `json:"severity" yaml:"severity"`
	// Recovery indicates given severity levels to use
	// this notification for recovery the give severity
	Recovery []sauron.Severity `json:"recovery" yaml:"recovery"`
}

// RawMessage wrapper both json and yaml raw message type
// (https://golang.org/pkg/encoding/json/#RawMessage)
// RawMessage is inspired by json.RawMessage but this is able
// to handle both JSON and YAML
type RawMessage struct {
	// for JSON
	data []byte
	// for YAML
	unmarshaler func(interface{}) error
	// codec indicates the codec
	// TODO: use iota
	codec string
}

// To converts raw message to the given type
func (m *RawMessage) To(v interface{}) error {
	switch m.codec {
	case "json":
		return json.Unmarshal(m.data, v)
	case "yaml":
		return m.unmarshaler(v)
	default:
		return fmt.Errorf("unknown codec %v", m.codec)
	}
}

// UnmarshalJSON sets data to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	m.codec = "json"
	m.data = append(m.data[0:0], data...)
	return nil
}

// UnmarshalYML implements the interface of yaml.Unmarshaler
func (m *RawMessage) UnmarshalYAML(unmarshaler func(interface{}) error) error {
	m.codec = "yaml"
	m.unmarshaler = unmarshaler
	return nil
}
