package main

import (
	"flag"
	"fmt"
	"net/http"

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

	logm := srvManager.LogMetric()

	if *env == "" {
		logm.Fatal("environment is empty")
	}

	if err := srvManager.Init(); err != nil {
		logm.Fatalf("fail to init services, err:%v", err)
	}

	// checks services
	ctrl := srvManager.Controller()
	etcdAPI := srvManager.Etcd()

	// start controller
	go func() {
		logm.Fatal(ctrl())
	}()

	// create core
	coreRoot := fmt.Sprintf(`/%s/pierce`, *env)
	coreMgr := core.NewCore(etcdAPI, coreRoot, logm)
	coreMgr.Start()

	// create socket handler
	if *enableSocket {
		skServer := socket.NewServer(coreMgr, logm)
		socketHandler, err := skServer.Create()
		if err != nil {
			logm.Fatal(err)
		}

		logm.Infof("[main] starting socket API server")
		http.Handle("/socket.io/", socketHandler)
	}

	// create rest handler
	if *enableRest {
		restHandler := rest.BuildHTTPHandler(logm)

		logm.Infof("[main] starting restful API server")
		http.Handle("/", restHandler)
	}
	logm.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
