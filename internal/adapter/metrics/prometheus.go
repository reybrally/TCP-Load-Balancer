package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/reybrally/TCP-Load-Balancer/internal/domain/port"
)

type PrometheusMetrics struct {
	connectionsTotal    *prometheus.CounterVec
	connectionsActive   *prometheus.GaugeVec
	connectionErrors    *prometheus.CounterVec
	connectionDuration  *prometheus.HistogramVec
	backendHealthStatus *prometheus.GaugeVec
	healthChecksTotal   *prometheus.CounterVec
}

func NewPrometheusMetrics() port.MetricsCollector {
	return &PrometheusMetrics{
		connectionsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tcp_lb_connections_total",
				Help: "Total number of TCP connections processed",
			},
			[]string{"backend"},
		),
		connectionsActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tcp_lb_connections_active",
				Help: "Number of currently active TCP connections",
			},
			[]string{"backend"},
		),
		connectionErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tcp_lb_connection_errors_total",
				Help: "Total number of connection errors",
			},
			[]string{"backend", "error_type"},
		),
		connectionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "tcp_lb_connection_duration_seconds",
				Help:    "Duration of TCP connections in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"backend"},
		),
		backendHealthStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tcp_lb_backend_healthy",
				Help: "Backend health status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"backend"},
		),
		healthChecksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "tcp_lb_health_checks_total",
				Help: "Total number of health checks performed",
			},
			[]string{"backend", "status"},
		),
	}
}

func (pm *PrometheusMetrics) IncConnectionsTotal(backend string) {
	pm.connectionsTotal.WithLabelValues(backend).Inc()
}

func (pm *PrometheusMetrics) IncConnectionsActive(backend string) {
	pm.connectionsActive.WithLabelValues(backend).Inc()
}

func (pm *PrometheusMetrics) DecConnectionsActive(backend string) {
	pm.connectionsActive.WithLabelValues(backend).Dec()
}

func (pm *PrometheusMetrics) IncConnectionErrors(backend string, errorType string) {
	pm.connectionErrors.WithLabelValues(backend, errorType).Inc()
}

func (pm *PrometheusMetrics) ObserveConnectionDuration(backend string, duration float64) {
	pm.connectionDuration.WithLabelValues(backend).Observe(duration)
}

func (pm *PrometheusMetrics) SetBackendHealthStatus(backend string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	pm.backendHealthStatus.WithLabelValues(backend).Set(value)
}

func (pm *PrometheusMetrics) IncHealthChecksTotal(backend string, status string) {
	pm.healthChecksTotal.WithLabelValues(backend, status).Inc()
}
