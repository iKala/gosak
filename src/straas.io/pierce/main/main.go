package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/base/metric"
	"straas.io/pierce/rest"
	"straas.io/service/common"
	"straas.io/service/manager"
)

var (
	portCtrl        = flag.Int("portCtrl", 8000, "port for health check")
	portRest        = flag.Int("portRest", 11300, "Restful API port")
	metricExportTag = flag.String("metricExportTag", "", "metric export tag")

	log        = logger.Get()
	srvManager = manager.New(common.MetricExporter)
)

func main() {
	flag.Parse()

	if err := srvManager.Init(); err != nil {
		log.Fatalf("fail to init services, err:%v", err)
	}

	// checks
	srvManager.MustGet(common.MetricExporter)

	go func() {
		log.Fatal(ctrl.RunController(*portCtrl))
	}()

	stat := metric.New("pierce")
	handler := rest.BuildHTTPHandler(log, stat)

	log.Infof("[main] starting restful API server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portRest), handler))
}
