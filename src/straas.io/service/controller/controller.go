package metric

import (
	"flag"

	"straas.io/base/ctrl"
	"straas.io/service/common"
)

func init() {
	common.Register(&service{})
}

type service struct {
	port int
}

func (s *service) Type() common.ServiceType {
	return common.Controller
}

func (s *service) AddFlags() {
	flag.IntVar(&s.port, "common.ctrl_port", 8000, "controller port")
}

func (s *service) New(get common.ServiceGetter) (interface{}, error) {
	return func() error {
		return ctrl.RunController(s.port)
	}, nil
}

func (s *service) Dependencies() []common.ServiceType {
	return nil
}
