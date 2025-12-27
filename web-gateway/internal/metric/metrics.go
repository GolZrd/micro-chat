package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "micro_chat"
	appName   = "web_gateway"
)

type Metrics struct {
	httpRequestTotal     *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
}

var metrics *Metrics

func Init() error {
	metrics = &Metrics{
		httpRequestTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      appName + "_request_total",
				Help:      "Количество HTTP запросов по методам и статусам",
			},
			[]string{"method", "path", "status"},
		),
		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      appName + "_request_duration_seconds",
				Help:      "Время выполнения HTTP запроса в секундах",
				Buckets:   []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
			}, []string{"path"},
		),
		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      appName + "_requests_in_flight",
				Help:      "Количество активных HTTP запросов",
			},
		),
	}

	return nil
}

func IncHttpRequestTotal(method string, path string, status string) {
	metrics.httpRequestTotal.WithLabelValues(method, path, status).Inc()
}

func ObserveHttpRequestDuration(path string, duration float64) {
	metrics.httpRequestDuration.WithLabelValues(path).Observe(duration)
}

func IncHttpRequestsInFlight() {
	metrics.httpRequestsInFlight.Inc()
}

func DecHttpRequestsInFlight() {
	metrics.httpRequestsInFlight.Dec()
}
