package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of requests received by status code.",
		},
		[]string{"path", "method", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets, 
		},
		[]string{"path", "method"},
	)


	inFlightRequests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of requests being processed.",
		},
		[]string{"path", "method"},
	)


	DependencyDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "app_dependency_duration_seconds",
			Help:    "Duration of outgoing requests to other services (Ping->Pong).",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target_service", "method"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(inFlightRequests)   
	prometheus.MustRegister(DependencyDuration) 
}

func PrometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		path := c.Path()
		if path == "" {
			path = c.Request().URL.Path
		}
		method := c.Request().Method

		inFlightRequests.WithLabelValues(path, method).Inc()


		defer inFlightRequests.WithLabelValues(path, method).Dec()

		err := next(c)

		status := c.Response().Status
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(path, method).Observe(duration)

		return err
	}
}

func MetricsHandler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.Handler())
}