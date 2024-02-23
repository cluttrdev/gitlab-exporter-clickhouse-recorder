package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler() http.Handler {
	return promhttp.Handler()
}
