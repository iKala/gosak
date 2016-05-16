package util

import (
	"straas.io/sauron"
)

func NewTestPlugin() *TestPlugin {
	return &TestPlugin{}
}

type TestPlugin struct {
	PluginName string
	RunFunc    func(ctx sauron.PluginContext) error
}

func (t *TestPlugin) Name() string {
	return t.PluginName
}

func (t *TestPlugin) Run(ctx sauron.PluginContext) error {
	return t.RunFunc(ctx)
}

func (t *TestPlugin) HelpMsg() string {
	return ""
}
