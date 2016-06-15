package rest

import (
	"fmt"
	"net/http"

	"straas.io/base/logger"
	"straas.io/base/rest"
)

// NewRest builds and returns a RESTful API handler
func NewRest(log logger.Logger) http.Handler {
	r := rest.NewRest(log)
	r.Route("GET", "/healthcheck", healthCheck)
	return r.GetHandler()
}

func healthCheck(w http.ResponseWriter, req *http.Request) *rest.Error {
	fmt.Fprintf(w, "OK")
	return nil
}
