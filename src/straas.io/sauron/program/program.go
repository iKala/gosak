package program

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"straas.io/base/logger"
	"straas.io/base/timeutil"
	"straas.io/sauron"
	"straas.io/sauron/core"
)

var (
	log = logger.Get()
)

// New creates a program instance
func New(
	envs []string,
	dryRun bool,
	jobPattern string,
	cfgMgr sauron.Config,
	tickInterval time.Duration) (sauron.Program, error) {

	// load jobs
	jobs, err := cfgMgr.LoadJobs(envs...)
	if err != nil {
		return nil, fmt.Errorf("[program] fail to load jobs, err:%v", err)
	}

	return &programImpl{
		envs:         envs,
		jobs:         jobs,
		cfgMgr:       cfgMgr,
		dryRun:       dryRun,
		tickInterval: tickInterval,
		jobPattern:   jobPattern,
	}, nil
}

type programImpl struct {
	cfgMgr        sauron.Config
	envs          []string
	dryRun        bool
	jobPattern    string
	tickInterval  time.Duration
	plugins       []sauron.Plugin
	jobs          []sauron.JobMeta
	eventHandlers []func(sauron.JobEvent)

	runner sauron.JobRunner
}

func (p *programImpl) AddPlugin(plugins ...sauron.Plugin) {
	// prepare plugins
	for _, plugin := range plugins {
		log.Infof("[program] register plugin %s", plugin.Name())
	}
	p.plugins = append(p.plugins, plugins...)
}

func (p *programImpl) AddEventHandler(handlers ...func(sauron.JobEvent)) {
	p.eventHandlers = append(p.eventHandlers, handlers...)
}

func (p *programImpl) Start() error {
	log.Info("[program] environemnt", p.envs)

	// prepare job runner
	if err := p.initRunner(); err != nil {
		return err
	}
	log.Info("[program] start to run jobs")

	go p.eventLoop()

	if !p.dryRun {
		p.runner.Start()
		return nil
	}
	return p.doDryRun()
}

func (p *programImpl) Stop() error {
	// not implement yet
	return nil
}

func (p *programImpl) eventLoop() {
	for e := range p.runner.Events() {
		for _, h := range p.eventHandlers {
			h(e)
		}
	}
}

func (p *programImpl) doDryRun() error {
	for _, j := range p.jobs {
		if p.jobPattern != "" && !strings.Contains(j.JobID, p.jobPattern) {
			continue
		}
		if err := p.runner.RunJob(j); err != nil {
			return err
		}
	}
	return nil
}

func (p *programImpl) initRunner() error {
	// runnerID
	runnerID := fmt.Sprintf("RN%d", rand.Int63())
	log.Infof("[program] Sauron runner id %s", runnerID)

	// create store
	statusStore, err := core.NewStore()
	if err != nil {
		return fmt.Errorf("[program] fail to init store, err:%v", err)
	}

	clock := timeutil.NewRealClock()
	// create output
	output := core.NewOutput(p.dryRun)
	// prepare engine factory
	engFactory := func() sauron.Engine {
		return core.NewEngine(statusStore, output)
	}
	// prepare ticker
	var ticker <-chan time.Time
	// dry run only need to tick once immediately
	if !p.dryRun {
		ticker = time.NewTicker(p.tickInterval).C
	}
	// create runner
	p.runner = core.NewJobRunner(
		runnerID,
		engFactory,
		ticker,
		statusStore,
		output,
		p.jobs,
		p.plugins,
		clock,
	)
	return nil
}
