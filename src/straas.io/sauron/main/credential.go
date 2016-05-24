package main

import (
	"straas.io/sauron/core"
)

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
	Scope []string `json:"scope" yaml:"scope"`
}

func loadCredential(cfgMgr core.Config, envs []string) (*Credential, error) {
	// load credential
	credential := &Credential{}
	if err = cfgMgr.LoadConfig("credential", credential); err != nil {
		return nil, err
	}
	return credential, nil
}
