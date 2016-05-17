package sauron

import (
	"time"
)

// define job event codes
const (
	EventConsErr  EventCode = iota
	EventTimeout  EventCode = iota
	EventParseErr EventCode = iota
)

// EventCode is type of job event code reported by JobRunner
type EventCode int32

// Severity defines severity level less value more severe
// a.k.a P0 is the most severe
type Severity int

// JobMeta represents a job
type JobMeta struct {
	// JobID is the job ID
	JobID string `json:"-" yaml:"-"`
	// Env is for environment (e.g. prod, staging)
	Env string `json:"-" yaml:"-"`
	// DryRun indicates whether run in dryrun mode
	DryRun bool `json:"-" yaml:"-"`
	// Script is script content
	Script string `json:"script" yaml:"script"`
	// Interval is running interval of the job
	Interval time.Duration `json:"interval" yaml:"interval"`
}

type FuncReturn func(ctx PluginContext) error

// PluginContext defines an interface for plugin to communicate
// with engine and job
type PluginContext interface {
	// JobMeta returns the job meta data
	JobMeta() JobMeta
	// ArgInt return the ith integer arguments of the plugin
	ArgInt(i int) (int64, error)
	// ArgFloat return the ith float arguments of the plugin
	ArgFloat(i int) (float64, error)
	// ArgString return the ith string arguments of the plugin
	ArgString(i int) (string, error)
	// ArgBoolean return the ith boolean arguments of the plugin
	ArgBoolean(i int) (bool, error)
	// ArgObject return the ith object argument of the plugin
	ArgObject(i int, v interface{}) error
	// CallFunction calls the ith arguments of the plugin as a javascript function
	CallFunction(i int, args ...interface{}) (interface{}, error)
	// IsCallable check if the ith argument is a function or not
	IsCallable(i int) bool
	// ArgLen return the number of arguments of the plugin
	ArgLen() int
	// Return sets the return value of the plugin
	Return(v interface{}) error
	// Store return the status store
	Store() Store
}

// Plugin defines the interface of javascript plugin func
type Plugin interface {
	// Name is the function name of the plugin
	Name() string
	// Run runs the plugin with a context
	Run(ctx PluginContext) error
	// HelpMsg returns the help msg of the plugin
	HelpMsg() string
}

// EngineFactory creates a script engine
type EngineFactory func() Engine

// Engine is the interface of script engine
type Engine interface {
	// SetJobMeta puts job metadata into engine
	SetJobMeta(meta JobMeta) error
	// AddPlugin registers plugins to engine
	AddPlugin(p Plugin) error
	// Run runs the engine with script in JobMeta
	Run() error
	// Get get the variable from vm
	Get(string) (interface{}, error)
	// Set set a variable to vm
	Set(string, interface{}) error
}

// Store is the abstract interface for status store
type Store interface {
	// Get returns data from store.
	Get(ns, key string, v interface{}) (bool, error)
	// Set puts data into store
	Set(ns, key string, v interface{}) error
	// Update updates data in store
	Update(ns, key string, v interface{},
		action func(interface{}) (interface{}, error)) error
}

// JobEvent define the JobRunner events
type JobEvent struct {
	// Code is the event code
	Code EventCode
	// JobID is the job id corresponding to the event
	JobID string
}

// JobRunner runner is responsible for invoking jobs according to their interval
// and record job status and result
type JobRunner interface {
	// Start start the runner
	Start() error
	// Stop stops the runner
	Stop() error
	// Update updates the jobs (from wating etcd or zk at runtime)
	Update(jobs []JobMeta) error
	// Events reports noticeable events of the runner
	Events() <-chan JobEvent
}

// Config manages the config
// TODO: not implement yet
type Config interface {
	// LoadJobs loads jobs of the given env
	LoadJobs(envs ...string) ([]JobMeta, error)
	// LoadCOnfig load config of the path
	LoadConfig(path string, v interface{}) error
	// AddChangeListener add config change listener
	AddChangeListener(func())
}
