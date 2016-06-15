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

// NewRest creates a new restful API builder
func NewRest(log logger.Logger) Rest {
	router := httprouter.New()
	n := negroni.Classic()
	n.UseHandler(router)

	return &restImpl{
		router:     router,
		middleware: n,
		log:        log,
	}
}

// Rest is the abstract interface to build restful API
// We will need to expand this interface if we need to use "Route Specific
// Middleware" in negroni or "Named Parameters" in httprouter.
type Rest interface {
	// GetHandler returns a http.Handler that can be passed to http.ListenAndServe
	GetHandler() http.Handler
	// Use registers a middleware function
	Use(fn middlewareFunc)
	// Route registers a route handler. When an error occurs, the handler should
	// just return an Error and let this package log and generate http error
	// response for you.
	Route(method, path string, handle handlerFunc)
}

type restImpl struct {
	router     *httprouter.Router
	middleware *negroni.Negroni
	log        logger.Logger
}

func (r *restImpl) GetHandler() http.Handler {
	return r.middleware
}

func (r *restImpl) Use(fn middlewareFunc) {
	r.middleware.Use(negroni.HandlerFunc(fn))
}

func (r *restImpl) Route(method, path string, handle handlerFunc) {
	r.router.Handler(method, path, handlerWrapper{handle, r.log})
}

type middlewareFunc func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc)

type handlerFunc func(rw http.ResponseWriter, req *http.Request) *Error

// Error is self defined error type
type Error struct {
	Error  error
	Detail string
	Code   int
}

// handlerWrapper manages all http error handling
type handlerWrapper struct {
	fn  func(http.ResponseWriter, *http.Request) *Error
	log logger.Logger
}

// ServeHTTP logs error Detail and generates http error response
func (hw handlerWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if restErr := hw.fn(w, r); restErr != nil {
		if restErr.Detail != "" {
			hw.log.Errorf("error detail: %s", restErr.Detail)
		}
		http.Error(w, restErr.Error.Error(), restErr.Code)
	}
}
