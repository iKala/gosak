package metric

import (
	"fmt"

	"straas.io/sauron"
)

func NewQuery() sauron.Plugin {
	return &metricQueryPlugin{}
}

type metricQueryPlugin struct {
}

func (p *metricQueryPlugin) Name() string {
	return "mquery"
}

func (p *metricQueryPlugin) Run(ctx sauron.PluginContext) error {
	v1, err := ctx.ArgInt(0)
	if err != nil {
		return err
	}
	v2, err := ctx.ArgInt(1)
	if err != nil {
		return err
	}
	ans := v1 + v2

	meta := ctx.JobMeta()
	ctx.Return(fmt.Sprintf("Run %s, %d + %d = %d", meta.JobID, v1, v2, ans))
	return nil
}

func (p *metricQueryPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}
