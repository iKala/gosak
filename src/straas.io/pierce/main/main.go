package main

import (
	"flag"
	"fmt"
	"net/http"

	"straas.io/base/logger"
	"straas.io/pierce/rest"

	"github.com/codegangsta/negroni"
)

var (
	port = flag.Int("port", 11300, "Restful API port")
	log  = logger.Get()
)

func main() {
	flag.Parse()

	var handler rest.Handler

	router := rest.BuildRouter(handler)
	n := negroni.Classic()
	n.UseHandler(router)

	log.Infof("[main] starting restful API server")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), n))
}
