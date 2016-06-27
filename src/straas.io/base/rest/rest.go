// Package rest represents a builder for RESTful API
// Copied from https://github.com/browny/goweb-scaffold with Viper and
// facebookgo packages removed as we do not use it now.
// For error handling, refer to:
// https://elithrar.github.io/article/http-handler-error-handling-revisited/
package rest

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"straas.io/base/logger"

	"github.com/codegangsta/negroni"
	"github.com/facebookgo/stats"
	"github.com/julienschmidt/httprouter"
)

// MiddlewareFunc defines middleware for your restful API
type MiddlewareFunc func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc)

// Params is just a interface compatable with httprouter.Params
type Params interface {
	ByName(name string) string
}

// HandlerFunc handles http requests and returns Error on failure.
type HandlerFunc func(rw http.ResponseWriter, req *http.Request, ps Params) *Error

// Error is self defined error type
type Error struct {
	Error  error
	Detail string
	Code   int
}

// New creates a new restful API builder
func New(log logger.Logger, stat stats.Client) Rest {
	return &restImpl{
		router:     httprouter.New(),
		middleware: negroni.New(negroni.NewRecovery(), &restLogger{Logger: log}),
		log:        log,
		stat:       stat,
	}
}

// Rest is the abstract interface to build restful API
// We will need to expand this interface if we need to use "Route Specific
// Middleware" in negroni
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
	router     *httprouter.Router
	middleware *negroni.Negroni
	log        logger.Logger
	stat       stats.Client
	once       sync.Once
}

func (r *restImpl) GetHandler() http.Handler {
	return r.middleware
}

func (r *restImpl) Use(fn MiddlewareFunc) {
	r.middleware.Use(negroni.HandlerFunc(fn))
}

func (r *restImpl) Route(method, path string, fn HandlerFunc) {
	r.router.Handle(method, path, r.wrapper(method, path, fn))
	r.once.Do(func() { r.middleware.UseHandler(r.router) })
}

func (r *restImpl) wrapper(method, path string, fn HandlerFunc) httprouter.Handle {
	// prepare metric name prefix
	// remove "/" at the begining of path
	// replace "/" as "."
	path = strings.TrimPrefix(path, "/")
	path = strings.Replace(path, "/", ".", -1)
	prefix := fmt.Sprintf("rest.%s.%s.", method, path)

	return func(rw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		defer r.stat.BumpTime(prefix + "proc_time").End()
		r.stat.BumpSum(prefix+"request", 1)
		if restErr := fn(rw, req, ps); restErr != nil {
			if restErr.Detail != "" {
				r.log.Errorf("error detail: %s", restErr.Detail)
			}
			http.Error(rw, restErr.Error.Error(), restErr.Code)
			r.stat.BumpSum(fmt.Sprintf("%ss%d", prefix, restErr.Code), 1)
			return
		}
	}
}

// restLogger wraps our own base/logger as a negroni middleware
type restLogger struct {
	logger.Logger
}

// ServeHTTP logs http access and process time
func (l *restLogger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	l.Printf("[rest] Started %s %s", r.Method, r.URL.Path)

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	l.Printf("[rest] Completed %v %s in %v", res.Status(), http.StatusText(res.Status()), time.Since(start))
}
