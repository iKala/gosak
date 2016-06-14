package main

import (
	"fmt"
	"net/http"

	"straas.io/pierce/rest"

	"github.com/codegangsta/negroni"
)

func main() {
	var handler rest.Handler

	router := rest.BuildRouter(handler)
	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(cors))
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%d", 3456))
}

// cors middleware (cross-origin resource sharing)
func cors(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	// Setting Access-Control-Allow-Origin to wildcard grants too much access.
	// Turn this on only if it is necessary.
	// rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, X-CSRF-Token, X-PINGOTHER")

	if req.Method == "OPTIONS" {
		method := req.Header.Get("Access-Control-Request-Method")

		if method == "" {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}

		rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		return
	}

	next(rw, req)
}
