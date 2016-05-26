package main

import (
	"straas.io/sauron"
)

// Config is main config of sauron
type Config struct {
	// Credential refers following struct
	Credential Credential `json:"credential" yaml:"credential"`
	// Env2GCP maps env to gcp proejcts
	Env2GCP map[string]string `json:"env2gcp" yaml:"env2gcp"`
}

// Credential defines necessary credential info for sauron
type Credential struct {
	// SlackToken is access token of slack
	SlackToken string `json:"slack_token" yaml:"slack_token"`
	// PagerDutyToken is access token of pagerduty
	PagerDutyToken string `json:"pagerduty_token" yaml:"pagerduty_token"`
	// GCP is a map from GCP project id to GCP credential
	GCP map[string]GCP `json:"gcp" yaml:"gcp"`
}

// GCP is gcp credential
type GCP struct {
	// Email is GCP service account
	Email string `json:"email" yaml:"email"`
	// PrivatKey is GCP private key of the service account
	PrivateKey string `json:"private_key" yaml:"private_key"`
	// Scope is GCP scope
	Scopes []string `json:"scopes" yaml:"scopes"`
}

func loadConfig(cfgMgr sauron.Config, envs []string) (Config, error) {
	// load credential
	cfg := Config{}
	if err := cfgMgr.LoadConfig("sauron", &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
