package util

import (
	"straas.io/sauron"
)

// NewTestPlugin creates a plugin for testing
// user can change implementation on demand
func NewTestPlugin() *TestPlugin {
	return &TestPlugin{}
}

// TestPlugin define a plugin for test
type TestPlugin struct {
	// PluginName for testre to replace plugin name
	PluginName string
	// Help for testre to replace help msg
	Help string
	// RunFunc for testre to replace run method
	RunFunc func(ctx sauron.PluginContext) error
}

// Name refer Plugin.Name
func (t *TestPlugin) Name() string {
	return t.PluginName
}

// Run refer Plugin.Run
func (t *TestPlugin) Run(ctx sauron.PluginContext) error {
	return t.RunFunc(ctx)
}

// HelpMsg refer Plugin.HelpMsg
func (t *TestPlugin) HelpMsg() string {
	return t.Help
}
