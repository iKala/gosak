package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/pierce/rest"
)

var (
	portCtrl = flag.Int("portCtrl", 8000, "port for health check")
	portRest = flag.Int("portRest", 11300, "Restful API port")
	log      = logger.Get()
)

func main() {
	flag.Parse()

	handler := rest.New(log)

	go func() {
		log.Fatal(ctrl.RunController(*portCtrl))
	}()

	log.Infof("[main] starting restful API server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portRest), handler))
}
