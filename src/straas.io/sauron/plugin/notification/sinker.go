package notification

import (
	"errors"

	"straas.io/sauron"
)

// SinkerFactory defines a type to create sinker
type SinkerFactory func() Sinker

// Sinker defines an interface to sink notifiations
type Sinker interface {
	// Name return sinker name
	Name() string
	// Sink sends notification
	Sink(config interface{}, severity sauron.Severity, recovery bool, desc string) error
	// ConfigFactory creates an instance to unmarshal sinker specific config,
	// the result will be passed as first argument in Sink method
	ConfigFactory() interface{}
}

// sinkerGroup wrapers all sinker for main process to treat as a single call
type sinkerGroup struct {
	sinkers []Sinker
	cfgs    []interface{}
}

func (g *sinkerGroup) empty() bool {
	return len(g.sinkers) == 0
}

func (g *sinkerGroup) sinkAll(ctx sauron.PluginContext, severity sauron.Severity,
	recovery bool, desc string) error {
	// for dryrun mode, print is enough
	for _, sinker := range g.sinkers {
		ctx.Infof("[notification] sink %s(severity:%d, recovery:%v, desc:%s)",
			sinker.Name(), severity, recovery, desc)
	}
	// dry cannot do folloing operations
	if ctx.JobMeta().DryRun {
		return nil
	}

	// TODO: keep retry until success or overwrite by later sinks
	errs := []error{}
	for i, s := range g.sinkers {
		if err := s.Sink(g.cfgs[i], severity, recovery, desc); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	// combine erros to an error
	msg := ""
	for _, err := range errs {
		msg += err.Error()
		msg += "\n"
	}
	return errors.New(msg)
}
