package pprof

import (
	"net/http"
	"net/http/pprof"
)

type Handler struct {
	http.ServeMux
}

func NewHandler() *Handler {
	h := &Handler{}

	h.HandleFunc("/", pprof.Index)
	h.HandleFunc("/cmdline", pprof.Cmdline)
	h.HandleFunc("/profile", pprof.Profile)
	h.HandleFunc("/symbol", pprof.Symbol)
	h.HandleFunc("/trace", pprof.Trace)

	return h
}

func Register(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/", pprof.Index)
	mux.HandleFunc(prefix+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(prefix+"/profile", pprof.Profile)
	mux.HandleFunc(prefix+"/symbol", pprof.Symbol)
	mux.HandleFunc(prefix+"/trace", pprof.Trace)
}
