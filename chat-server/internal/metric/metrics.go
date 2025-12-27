package metric

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "micro_chat"
	appName   = "chat-server"
)

type Metrics struct {
	requestTotal     *prometheus.CounterVec
	requestDuration  *prometheus.HistogramVec
	requestsInFlight prometheus.Gauge
}

var metrics *Metrics

func Init(_ context.Context) error {
	metrics = &Metrics{
		requestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      appName + "_request_total",
				Help:      "Количество запросов к серверу по методам и статусам",
			},
			[]string{"status", "method"},
		),
		requestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      appName + "_request_duration_seconds",
				Help:      "Время выполнения запроса к серверу в секундах",
				Buckets:   []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
			}, []string{"method"},
		),
		requestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "grpc",
				Name:      appName + "_requests_in_flight",
				Help:      "Количество активных запросов к серверу",
			},
		),
	}

	return nil
}

func IncRequestTotal(status string, method string) {
	metrics.requestTotal.WithLabelValues(status, method).Inc()
}

func ObserveRequestDuration(method string, duration float64) {
	metrics.requestDuration.WithLabelValues(method).Observe(duration)
}

func IncRequestsInFlight() {
	metrics.requestsInFlight.Inc()
}

func DecRequestsInFlight() {
	metrics.requestsInFlight.Dec()
}
