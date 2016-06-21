package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/base/metric"
	"straas.io/external/fluent"
	"straas.io/pierce/rest"
)

var (
	portCtrl        = flag.Int("portCtrl", 8000, "port for health check")
	portRest        = flag.Int("portRest", 11300, "Restful API port")
	fluentEnable    = flag.Bool("fluentEnable", false, "fluent enable")
	fluentHost      = flag.String("fluentHost", "", "fluent hostname")
	fluentPort      = flag.Int("fluentPort", 24224, "fluent port")
	metricExportTag = flag.String("metricExportTag", "", "metric export tag")

	log = logger.Get()
)

func main() {
	flag.Parse()

	fluentLogger, err := fluent.New(*fluentEnable, *fluentHost, *fluentPort)
	if err != nil {
		log.Fatalf("fail to create fluent, err:%v", err)
	}

	metric.StartExport(fluentLogger, *metricExportTag, nil)

	go func() {
		log.Fatal(ctrl.RunController(*portCtrl))
	}()

	stat := metric.New("pierce")
	handler := rest.BuildHTTPHandler(log, stat)

	log.Infof("[main] starting restful API server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portRest), handler))
}
