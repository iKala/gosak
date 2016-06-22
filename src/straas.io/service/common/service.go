package common

import (
	"fmt"
)

// define service types
const (
	Fluent         ServiceType = "fluent"
	MetricExporter ServiceType = "metric_exporter"
)

// ServiceType define service type
type ServiceType string

// ServiceGetter defines a func type to get dependent service by type
type ServiceGetter func(ServiceType) interface{}

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
