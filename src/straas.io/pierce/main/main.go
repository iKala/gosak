package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/external"
	"straas.io/pierce/core"
	"straas.io/pierce/rest"
	"straas.io/pierce/socket"
	"straas.io/service/common"
	"straas.io/service/manager"
)

var (
	port         = flag.Int("pierce.port", 11300, "Restful API port")
	enableSocket = flag.Bool("pierce.enable_socket", true, "enable socket or not")
	enableRest   = flag.Bool("pierce.enable_rest", true, "enable rest or not")
	env          = flag.String("pierce.env", "", "running environment")

	// service manager
	srvManager = manager.New("pierce",
		common.Controller,
		common.Etcd,
		common.MetricExporter,
	)
)

func main() {
	flag.Parse()

	log := srvManager.Logger()

	if *env == "" {
		log.Fatal("environment is empty")
	}

	if err := srvManager.Init(); err != nil {
		log.Fatalf("fail to init services, err:%v", err)
	}

	// checks services
	ctrl := srvManager.MustGet(common.Controller).(func() error)
	etcdAPI := srvManager.MustGet(common.Etcd).(external.Etcd)

	// start contruller
	go func() {
		log.Fatal(ctrl())
	}()

	// create core
	coreMgr := core.NewCore(etcdAPI, "/pierce")
	coreMgr.Start()

	// create socket handler
	if *enableSocket {
		skServer := socket.NewServer(coreMgr)
		socketHandler, err := skServer.Create()
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("[main] starting socket API server")
		http.Handle("/socket.io/", socketHandler)
	}

	// create rest handler
	if *enableRest {
		restHandler := rest.BuildHTTPHandler(log, srvManager.Metric())

		log.Infof("[main] starting restful API server")
		http.Handle("/", restHandler)
	}
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

//./main -pierce.env=test -common.etcd_urls="http://192.168.1.57:2379"
