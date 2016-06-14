package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/base/ctrl"
	"straas.io/base/logger"
	"straas.io/pierce/rest"

	"github.com/codegangsta/negroni"
)

var (
	portCtrl = flag.Int("portCtrl", 8000, "port for health check")
	portRest = flag.Int("portRest", 11300, "Restful API port")
	log      = logger.Get()
)

func main() {
	flag.Parse()

	var handler rest.Handler

	router := rest.BuildRouter(handler)
	n := negroni.Classic()
	n.UseHandler(router)

	go func() {
		log.Fatal(ctrl.RunController(*portCtrl))
	}()

	log.Infof("[main] starting restful API server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portRest), n))
}
