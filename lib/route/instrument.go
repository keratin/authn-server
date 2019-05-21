package route

import (
	"net/http"
	"strconv"

	"github.com/felixge/httpsnoop"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "How many HTTP requests processed, partitioned by name and status code",
		},
		[]string{"name", "code"},
	)
	httpTimings = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_times",
			Help:    "The duration of HTTP requests, partitioned by name",
			Buckets: prometheus.ExponentialBuckets(0.001, 10, 5),
		},
		[]string{"name"},
	)
)

func init() {
	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(httpTimings)
}

func InstrumentRoute(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics := httpsnoop.CaptureMetrics(next, w, r)
		httpRequests.WithLabelValues(name, strconv.Itoa(metrics.Code)).Inc()
		httpTimings.WithLabelValues(name).Observe(float64(metrics.Duration.Seconds()))
	})
}
