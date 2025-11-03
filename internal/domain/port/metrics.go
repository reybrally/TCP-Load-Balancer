package port

type MetricsCollector interface {
	IncConnectionsTotal(backend string)

	IncConnectionsActive(backend string)

	DecConnectionsActive(backend string)

	IncConnectionErrors(backend string, errorType string)

	ObserveConnectionDuration(backend string, duration float64)

	SetBackendHealthStatus(backend string, healthy bool)

	IncHealthChecksTotal(backend string, status string)
}
