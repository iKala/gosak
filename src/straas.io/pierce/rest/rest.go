package rest

import (
	"fmt"
	"net/http"

	"github.com/facebookgo/stats"

	"straas.io/base/logger"
	"straas.io/base/rest"
)

// BuildHTTPHandler builds and returns a RESTful API handler that can be passed
// to http.ListenAndServe
func BuildHTTPHandler(log logger.Logger, stat stats.Client) http.Handler {
	r := rest.New(log, stat)
	r.Route("GET", "/healthcheck", healthCheck)
	return r.GetHandler()
}

func healthCheck(w http.ResponseWriter, req *http.Request) *rest.Error {
	fmt.Fprintf(w, "OK")
	return nil
}
