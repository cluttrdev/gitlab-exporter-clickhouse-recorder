package pprof

import (
	"net/http"
	"net/http/pprof"
)

type handler struct {
	http.ServeMux
}

func NewHandler() http.Handler {
	h := &handler{}

	h.HandleFunc("/", pprof.Index)
	h.HandleFunc("/cmdline", pprof.Cmdline)
	h.HandleFunc("/profile", pprof.Profile)
	h.HandleFunc("/symbol", pprof.Symbol)
	h.HandleFunc("/trace", pprof.Trace)

	return h
}
