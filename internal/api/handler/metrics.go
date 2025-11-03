package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/reybrally/TCP-Load-Balancer/internal/pkg/logger"
)

type MetricsHandler struct {
	logger *logger.Logger
}

func NewMetricsHandler(logger *logger.Logger) *MetricsHandler {
	return &MetricsHandler{
		logger: logger,
	}
}

func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Debugf("Metrics endpoint accessed from %s", r.RemoteAddr)
	promhttp.Handler().ServeHTTP(w, r)
}

func StartMetricsServer(port string, logger *logger.Logger) error {
	mux := http.NewServeMux()

	metricsHandler := NewMetricsHandler(logger)
	mux.Handle("/metrics", metricsHandler)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Infof("Metrics server starting on http://0.0.0.0:%s/metrics", port)

	return http.ListenAndServe(":"+port, mux)
}
