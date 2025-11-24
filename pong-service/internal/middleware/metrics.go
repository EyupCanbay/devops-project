package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Senior Notu 1: Global değişkenler yerine Struct yapısı
// Bu sayede test yazarken mock'layabilirsin ve değişken kirliliğini önlersin.
type PrometheusMetrics struct {
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	ResponseSize       *prometheus.HistogramVec // Ağ yükünü görmek için
	InFlightRequests   *prometheus.GaugeVec
	DependencyDuration *prometheus.HistogramVec
}

// Global instance (Handler'lardan erişmek için)
var Metrics *PrometheusMetrics

func InitMetrics() {
	Metrics = &PrometheusMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "myapp",          // Tüm uygulamalarımız bu çatı altında
				Subsystem: "pong_service",   // Servis adı burada ayrışır
				Name:      "http_requests_total",
				Help:      "Number of requests received by status code.",
			},
			[]string{"path", "method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "pong_service",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				// Senior Notu 2: Pong genellikle hızlıdır, daha hassas bucketlar koyabiliriz.
				Buckets: []float64{0.01, 0.05, 0.1, 0.2, 0.5, 1}, 
			},
			[]string{"path", "method"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "pong_service",
				Name:      "http_response_size_bytes",
				Help:      "Size of HTTP responses in bytes.",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 6), // 100B, 1KB, 10KB...
			},
			[]string{"path", "method"},
		),
		InFlightRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "myapp",
				Subsystem: "pong_service",
				Name:      "http_requests_in_flight",
				Help:      "Current number of requests being processed.",
			},
			[]string{"path", "method"},
		),
		DependencyDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "pong_service",
				Name:      "dependency_duration_seconds",
				Help:      "Duration of outgoing requests (Pong->Ping Ack).",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"target_service", "method"},
		),
	}

	// Metrikleri Kaydet
	prometheus.MustRegister(Metrics.RequestsTotal)
	prometheus.MustRegister(Metrics.RequestDuration)
	prometheus.MustRegister(Metrics.ResponseSize)
	prometheus.MustRegister(Metrics.InFlightRequests)
	prometheus.MustRegister(Metrics.DependencyDuration)
}

func PrometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Senior Notu 3: High Cardinality Koruması
		// Eğer path boş gelirse veya ID içeriyorsa, metrikleri patlatmaması için önlem.
		path := c.Path()
		if path == "" {
			path = "unknown_route"
		}
		method := c.Request().Method

		// İşlem başladı (Gauge arttır)
		Metrics.InFlightRequests.WithLabelValues(path, method).Inc()
		defer Metrics.InFlightRequests.WithLabelValues(path, method).Dec()

		err := next(c)

		status := c.Response().Status
		if err != nil {
			c.Error(err)
			status = c.Response().Status
		}

		duration := time.Since(start).Seconds()
		respSize := float64(c.Response().Size) // Response boyutunu al

		// Metrikleri işle
		Metrics.RequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		Metrics.RequestDuration.WithLabelValues(path, method).Observe(duration)
		Metrics.ResponseSize.WithLabelValues(path, method).Observe(respSize)

		return err
	}
}

func MetricsHandler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.Handler())
}