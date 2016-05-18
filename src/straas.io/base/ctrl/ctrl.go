package ctrl

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

// RunController run controller service
func RunController(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// OK
		fmt.Fprintln(w, "OK")
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}
