package common

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"straas.io/base/logmetric"
	"straas.io/external"
)

// define service types, pls list in alphabetical order
const (
	Controller     ServiceType = "ctrl"
	Etcd           ServiceType = "etcd"
	Fluent         ServiceType = "fluent"
	MetricExporter ServiceType = "metric_exporter"
	MySQL          ServiceType = "mysql"
)

// ServiceType define service type
type ServiceType string

// ServiceGetter defines an interface to get dependent service by type
type ServiceGetter interface {
	// MustGet returns the serivce instance and panic if any error
	MustGet(ServiceType) interface{}
	// Get return the service instance
	Get(ServiceType) (interface{}, error)
	// LogMetric returns logger and metric
	LogMetric() logmetric.LogMetric
	// Controller returns controller start func
	Controller() func() error
	// MetricExporter return metric exporter stop func
	MetricExporter() func()
	// MySQL returns a MySQL client wrapped with gorm
	MySQL() *gorm.DB
	// Fluent return a fluent client
	Fluent() external.Fluent
	// Etcd return a etcd client
	Etcd() external.Etcd
}

// Service defines an interface for common services
type Service interface {
	// Type return serive type
	Type() ServiceType
	// AddFlags adds necessary flags
	AddFlags()
	// Dependencies returns types of dependent services
	Dependencies() []ServiceType
	// New creates an instance of the service
	New(ServiceGetter) (interface{}, error)
}

var (
	services = map[ServiceType]Service{}
)

// Register must be put in init()
func Register(s Service) {
	if _, ok := services[s.Type()]; ok {
		panic(fmt.Errorf("already register service %v", s.Type()))
	}
	for _, st := range s.Dependencies() {
		if st == s.Type() {
			panic(fmt.Errorf("cannot depend on itself"))
		}
		if _, ok := services[st]; !ok {
			continue
		}
		// mutual dependencies leads to cycle dependencies
		if dependOn(st, s.Type()) {
			panic(fmt.Errorf("cycle dependecies"))
		}
	}
	services[s.Type()] = s
}

// Services returns registered services
func Services() map[ServiceType]Service {
	return services
}

// dependOn checks if there are cyclic dependencies
func dependOn(s, t ServiceType) bool {
	touched := map[ServiceType]bool{}
	stack := []ServiceType{s}

	for len(stack) > 0 {
		// pop
		st := stack[0]
		stack = stack[1:]
		if st == t {
			return true
		}
		if touched[st] {
			continue
		}
		touched[st] = true
		// push
		ss, ok := services[st]
		if !ok {
			continue
		}
		stack = append(stack, ss.Dependencies()...)
	}
	return false
}
