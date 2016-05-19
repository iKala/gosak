package notification

import (
	"straas.io/base/encoding"
	"straas.io/sauron"
)

// Config is the root of notification configuration
type Config struct {
	Groups []*Group `json:"groups" yaml:"groups"`
}

// Group defines the notification group
type Group struct {
	// Name is group name
	Name string `json:"name" yaml:"name"`
	// Desc is group description
	Desc string `json:"desc" yaml:"desc"`
	// Notifications are a list of raw sinks
	RawSinkers []*encoding.RawMessage `json:"sinkers" yaml:"sinkers"`
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
