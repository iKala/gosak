package etcd

import (
	"errors"
	"flag"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/etcd/client"

	exetcd "straas.io/external/etcd"
	"straas.io/service/common"
)

const (
	keepAlive = 30 * time.Second
	timeout   = 10 * time.Second
)

func init() {
	common.Register(&service{})
}

type service struct {
	urls string
}

func (s *service) Type() common.ServiceType {
	return common.Etcd
}

func (s *service) AddFlags() {
	flag.StringVar(&s.urls, "common.etcd_urls", "", "metric fluent tag")
}

func (s *service) New(get common.ServiceGetter) (interface{}, error) {
	if s.urls == "" {
		return nil, errors.New("empty etcd host urls")
	}

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: keepAlive,
		}).Dial,
	}
	cfg := client.Config{
		Endpoints:               strings.Split(s.urls, ","),
		Transport:               transport,
		HeaderTimeoutPerRequest: timeout,
	}
	cli, err := client.New(cfg)
	if err != nil {
		return nil, err
	}
	return exetcd.NewEtcd(cli, timeout, get.LogMetric()), nil
}

func (s *service) Dependencies() []common.ServiceType {
	return nil
}
