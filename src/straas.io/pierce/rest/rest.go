package rest

import (
	"fmt"
	"net/http"

	"straas.io/base/logger"
	"straas.io/base/rest"
)

// New builds and returns a RESTful API handler that can be passed to http.ListenAndServe
func New(log logger.Logger) http.Handler {
	r := rest.New(log)
	r.Route("GET", "/healthcheck", healthCheck)
	return r.GetHandler()
}

func healthCheck(w http.ResponseWriter, req *http.Request) *rest.Error {
	fmt.Fprintf(w, "OK")
	return nil
}
