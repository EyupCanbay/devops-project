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
			Help: "Number of requests received.",
		},
		[]string{"path", "method", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_second",
			Help:    "HTTP request duration",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func InitMetrics() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

func PrometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		err := next(c)
		status := c.Response().Status
		duration := time.Since(start).Seconds()
		path := c.Request().URL.Path
		method := c.Request().Method

		httpRequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		httpRequestDuration.WithLabelValues(path, method).Observe(duration)
		return err
	}
}

func MetricsHandler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.Handler())
}
