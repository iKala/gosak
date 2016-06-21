package manager

import (
	"fmt"

	"straas.io/service/common"
	// services
	_ "straas.io/service/fluent"
	_ "straas.io/service/metric"
)

// New creates an instance of manager
func New(types ...common.ServiceType) Manager {
	return newMgr(common.Services(), types...)
}

func newMgr(services map[common.ServiceType]common.Service,
	types ...common.ServiceType) Manager {
	if len(types) == 0 {
		panic(fmt.Errorf("no service types"))
	}
	m := &managerImpl{
		services:  services,
		types:     types,
		instances: map[common.ServiceType]interface{}{},
	}
	// add flags
	touched := map[common.ServiceType]bool{}
	m.addFlags(touched, types...)
	return m
}

// Manager defines an interface for service manager
type Manager interface {
	// Init creates serivce instances according to flag and dependencies
	Init() error
	// MustGet returns the serivce instance and panic if any error
	MustGet(common.ServiceType) interface{}
	// Get return the service instance
	Get(common.ServiceType) (interface{}, error)
}

// some implementation throws panic bcz user misuses this manager
type managerImpl struct {
	inited    bool
	types     []common.ServiceType
	instances map[common.ServiceType]interface{}
	services  map[common.ServiceType]common.Service
}

func (m *managerImpl) Init() error {
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
	ts ...common.ServiceType) {

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
		m.addFlags(touched, srv.Dependencies()...)
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
	inst, err := s.New(m.MustGet)
	if err != nil {
		return err
	}
	m.instances[t] = inst
	return nil
}
