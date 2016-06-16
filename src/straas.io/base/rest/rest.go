// Package rest represents a builder for RESTful API
// Copied from https://github.com/browny/goweb-scaffold with Viper and
// facebookgo packages removed as we do not use it now.
// For error handling, refer to:
// https://elithrar.github.io/article/http-handler-error-handling-revisited/
package rest

import (
	"net/http"

	"straas.io/base/logger"

	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
)

// MiddlewareFunc defines middleware for your restful API
type MiddlewareFunc func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc)

// HandlerFunc handles http requests and returns Error on failure.
type HandlerFunc func(rw http.ResponseWriter, req *http.Request) *Error

// Error is self defined error type
type Error struct {
	Error  error
	Detail string
	Code   int
}

// New creates a new restful API builder
func New(log logger.Logger) Rest {
	return &restImpl{
		router:        httprouter.New(),
		middleware:    negroni.Classic(),
		log:           log,
		routerApplied: false,
	}
}

// Rest is the abstract interface to build restful API
// We will need to expand this interface if we need to use "Route Specific
// Middleware" in negroni or "Named Parameters" in httprouter.
type Rest interface {
	// GetHandler returns a http.Handler that can be passed to http.ListenAndServe
	GetHandler() http.Handler
	// Use registers a middleware function
	Use(fn MiddlewareFunc)
	// Route registers a route handler. When an error occurs, the handler should
	// just return an Error and let this package log and generate http error
	// response for you.
	Route(method, path string, handle HandlerFunc)
}

type restImpl struct {
	router        *httprouter.Router
	middleware    *negroni.Negroni
	log           logger.Logger
	routerApplied bool
}

func (r *restImpl) GetHandler() http.Handler {
	return r.middleware
}

func (r *restImpl) Use(fn MiddlewareFunc) {
	r.middleware.Use(negroni.HandlerFunc(fn))
}

func (r *restImpl) Route(method, path string, fn HandlerFunc) {
	r.router.HandlerFunc(method, path, wrapper(fn, r.log))
	if !r.routerApplied {
		r.middleware.UseHandler(r.router)
		r.routerApplied = true
	}
}

func wrapper(fn HandlerFunc, log logger.Logger) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		if restErr := fn(rw, req); restErr != nil {
			if restErr.Detail != "" {
				log.Errorf("error detail: %s", restErr.Detail)
			}
			http.Error(rw, restErr.Error.Error(), restErr.Code)
		}
	}
}
