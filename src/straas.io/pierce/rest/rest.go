package rest

import (
	"fmt"
	"net/http"

	"straas.io/base/logmetric"
	"straas.io/base/rest"
)

// BuildHTTPHandler builds and returns a RESTful API handler that can be passed
// to http.ListenAndServe
func BuildHTTPHandler(logm logmetric.LogMetric) http.Handler {
	r := rest.New(logm)
	r.Route("GET", "/healthcheck", healthCheck)
	return r.GetHandler()
}

func healthCheck(w http.ResponseWriter, req *http.Request, _ rest.Params) *rest.Error {
	fmt.Fprintf(w, "OK")
	return nil
}
