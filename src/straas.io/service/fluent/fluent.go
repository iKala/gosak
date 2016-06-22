package fluent

import (
	"flag"

	"straas.io/external/fluent"
	"straas.io/service/common"
)

func init() {
	common.Register(&service{})
}

type service struct {
	host   string
	port   int
	enable bool
}

func (s *service) Type() common.ServiceType {
	return common.Fluent
}

func (s *service) AddFlags() {
	flag.StringVar(&s.host, "common_fluent_host", "", "fluent host address")
	flag.IntVar(&s.port, "common_fluent_port", 24224, "fluent port")
	flag.BoolVar(&s.enable, "common_fluent_enable", false, "whether enable fluent")
}

func (s *service) New(common.ServiceGetter) (interface{}, error) {
	return fluent.New(s.enable, s.host, s.port)
}

func (s *service) Dependencies() []common.ServiceType {
	return nil
}
