package notification

import (
	"fmt"

	"straas.io/base/logger"
	"straas.io/sauron"
)

var (
	sinkerFac = map[string]SinkerFactory{}
	log       = logger.Get()
)

// TODO: add logs
// notify(<GROUP>)(<SEVERITY>, <RECOVERY>, <DESC>)
// e.g. notify("backend")(0, false, "Incident aaa.bbb 10 > 5")

// NewNotification creates notifidation plugin
func NewNotification(cfgMgr sauron.Config) (sauron.Plugin, error) {
	// TODO: handle update
	// now for simple, only load config here
	p := &notificationPlugin{}
	if err := p.loadConfig(cfgMgr); err != nil {
		return nil, err
	}
	return p, nil
}

// RegisterSinker registers notification sinker
func RegisterSinker(name string, factory SinkerFactory) error {
	_, ok := sinkerFac[name]
	if ok {
		return fmt.Errorf("%s already registered", name)
	}
	sinkerFac[name] = factory
	return nil
}

type notificationPlugin struct {
	groupInfoMap map[string]*groupInfo
}

type groupInfo struct {
	baseCfgs   []*BaseSinkCfg
	sinkerCfgs []interface{}
	sinkers    []Sinker
}

func (p *notificationPlugin) Name() string {
	return "notify"
}

// Run run the lastfor,
func (p *notificationPlugin) Run(ctx sauron.PluginContext) error {
	group, err := ctx.ArgString(0)
	if err != nil {
		return err
	}
	// check group
	info, ok := p.groupInfoMap[group]
	if !ok {
		return fmt.Errorf("unknown group %s", group)
	}
	// return callback function
	ctx.Return(p.genCallback(info))
	return nil
}

func (p *notificationPlugin) HelpMsg() string {
	return "<NO HELP MSG>"
}

func (p *notificationPlugin) loadConfig(cfgMgr sauron.Config) error {
	config := &Config{}
	if err := cfgMgr.LoadConfig("notification/config", config); err != nil {
		return fmt.Errorf("fail to load config, err:%v", err)
	}

	p.groupInfoMap = map[string]*groupInfo{}
	for _, g := range config.Groups {
		info := &groupInfo{}
		for _, n := range g.RawSinkers {
			base := &BaseSinkCfg{}
			if err := n.To(base); err != nil {
				return fmt.Errorf("fail to parse notificatio, err:%v", err)
			}
			// create sinker according to type
			sinker, err := newSinker(base.Type)
			if err != nil {
				return err
			}
			// parse sinker specific config from raw config
			sinkerCfg := sinker.ConfigFactory()
			if err := n.To(sinkerCfg); err != nil {
				return fmt.Errorf("fail to parse sinker specifi config, err:%v", err)
			}
			// add to info
			info.baseCfgs = append(info.baseCfgs, base)
			info.sinkerCfgs = append(info.sinkerCfgs, sinkerCfg)
			info.sinkers = append(info.sinkers, sinker)
		}

		p.groupInfoMap[g.Name] = info
	}
	return nil
}

func (p *notificationPlugin) genCallback(info *groupInfo) sauron.FuncReturn {
	return sauron.FuncReturn(func(ctx sauron.PluginContext) error {
		s, err := ctx.ArgInt(0)
		if err != nil {
			return err
		}
		resolve, err := ctx.ArgBoolean(1)
		if err != nil {
			return err
		}
		desc, err := ctx.ArgString(2)
		if err != nil {
			return err
		}
		sv := sauron.Severity(s)

		// get sinker group
		sgroup := newSinkerGroup(info, sv, resolve)
		if sgroup.empty() {
			return fmt.Errorf("no sinkers for severity:%v, resolve:%v", sv, resolve)
		}
		if err := sgroup.sinkAll(ctx.JobMeta().DryRun, sv, resolve, desc); err != nil {
			// TODO: log here
			// Tend to not return error here to affet alert main process
			// but not sure how to do
			ctx.Return(false)
			return nil
		}
		// success
		return ctx.Return(true)
	})
}

// newSinker create sinker the instance according to the given type
func newSinker(sinkerType string) (Sinker, error) {
	// TODO: use register ?
	fac, ok := sinkerFac[sinkerType]
	if !ok {
		return nil, fmt.Errorf("unknown notification type %s", sinkerType)
	}
	return fac(), nil
}

// newSinkerGroup finds sinkers match the given criteria
func newSinkerGroup(info *groupInfo, sv sauron.Severity, resolve bool) *sinkerGroup {
	sg := &sinkerGroup{}
	for i, bCfg := range info.baseCfgs {
		svs := bCfg.Severity
		if resolve {
			svs = bCfg.Recovery
		}
		if contains(svs, sv) {
			sg.sinkers = append(sg.sinkers, info.sinkers[i])
			sg.cfgs = append(sg.cfgs, info.sinkerCfgs[i])
		}
	}
	return sg
}

// contains checks if severity slices contain the target severity level
func contains(svs []sauron.Severity, sv sauron.Severity) bool {
	for _, s := range svs {
		if sv == s {
			return true
		}
	}
	return false
}
