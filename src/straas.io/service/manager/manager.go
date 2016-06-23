package manager

import (
	"flag"
	"fmt"

	"github.com/facebookgo/stats"

	"straas.io/base/logger"
	"straas.io/base/metric"
	"straas.io/service/common"
	// services
	_ "straas.io/service/controller"
	_ "straas.io/service/etcd"
	_ "straas.io/service/fluent"
	_ "straas.io/service/metric"
)

var (
	log = logger.Get()
)

// New creates an instance of manager
func New(moduleName string, types ...common.ServiceType) Manager {
	return newMgr(common.Services(), moduleName, types)
}

func newMgr(services map[common.ServiceType]common.Service,
	moduleName string, types []common.ServiceType) Manager {
	if len(types) == 0 {
		panic(fmt.Errorf("no service types"))
	}
	m := &managerImpl{
		services:  services,
		types:     types,
		stat:      metric.New(moduleName),
		instances: map[common.ServiceType]interface{}{},
	}
	// add flags
	touched := map[common.ServiceType]bool{}
	m.addFlags(touched, types)
	return m
}

// Manager defines an interface for service manager
type Manager interface {
	common.ServiceGetter
	// Init creates serivce instances according to flag and dependencies
	Init() error
}

// some implementation throws panic bcz user misuses this manager
type managerImpl struct {
	inited    bool
	types     []common.ServiceType
	stat      stats.Client
	instances map[common.ServiceType]interface{}
	services  map[common.ServiceType]common.Service
}

func (m *managerImpl) Init() error {
	if !flag.Parsed() {
		panic("flag is not parsed")
	}
	if m.inited {
		panic("already inited")
	}

	m.inited = true
	// Run DFS to build
	for _, t := range m.types {
		if err := m.build(t); err != nil {
			return err
		}
	}
	return nil
}

func (m *managerImpl) Logger() logger.Logger {
	return log
}

func (m *managerImpl) Metric() stats.Client {
	return m.stat
}

func (m *managerImpl) MustGet(t common.ServiceType) interface{} {
	s, err := m.Get(t)
	if err != nil {
		panic(err)
	}
	return s
}

func (m *managerImpl) Get(t common.ServiceType) (interface{}, error) {
	inst, ok := m.instances[t]
	if !ok {
		return nil, fmt.Errorf("fail to find service for %v", t)
	}
	return inst, nil
}

func (m *managerImpl) addFlags(touched map[common.ServiceType]bool,
	ts []common.ServiceType) {

	for _, t := range ts {
		if touched[t] {
			continue
		}
		touched[t] = true
		srv, ok := m.services[t]
		if !ok {
			panic(fmt.Errorf("serivce %v is not registered yet", t))
		}
		srv.AddFlags()
		m.addFlags(touched, srv.Dependencies())
	}
}

func (m *managerImpl) build(t common.ServiceType) error {
	// already built
	if _, ok := m.instances[t]; ok {
		return nil
	}
	s := m.services[t]
	// build dependencies first
	for _, st := range s.Dependencies() {
		if err := m.build(st); err != nil {
			return err
		}
	}
	// create flags for the service
	inst, err := s.New(m)
	if err != nil {
		return err
	}
	m.instances[t] = inst
	return nil
}
