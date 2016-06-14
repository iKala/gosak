// Package rest represents the REST layer
// Copied from https://github.com/browny/goweb-scaffold with Viper and
// facebookgo packages removed as we do not use it now.
package rest

import (
	"fmt"
	"net/http"

	"straas.io/base/logger"

	"github.com/julienschmidt/httprouter"
)

var log = logger.Get()

// Error is self defined error type
type Error struct {
	Error  error
	Detail string
	Code   int
}

// handlerWrapper manages all http error handling
type handlerWrapper func(http.ResponseWriter, *http.Request) *Error

func (fn handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if restErr := fn(w, r); restErr != nil {
		if restErr.Detail != "" {
			log.Errorf("error detail: %s", restErr.Detail)
		}
		http.Error(w, restErr.Error.Error(), restErr.Code)
	}
}

// Handler includes all http handler methods
type Handler struct {
}

// HealthCheck is called by Google cloud to do health check
func (rest *Handler) HealthCheck(w http.ResponseWriter, req *http.Request) *Error {
	fmt.Fprintf(w, "OK")
	return nil
}

// BuildRouter registers all routes
func BuildRouter(h Handler) *httprouter.Router {
	router := httprouter.New()

	router.Handler("GET", "/healthcheck", handlerWrapper(h.HealthCheck))

	return router
}
