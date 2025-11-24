package middleware

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Senior Notu 1: Global değişkenler yerine bir struct içine toplamak daha temizdir.
// İleride Dependency Injection yapabilirsin.
type PrometheusMetrics struct {
	RequestsTotal      *prometheus.CounterVec
	RequestDuration    *prometheus.HistogramVec
	RequestSize        *prometheus.HistogramVec // Request boyutu
	ResponseSize       *prometheus.HistogramVec // Response boyutu
	InFlightRequests   *prometheus.GaugeVec
	DependencyDuration *prometheus.HistogramVec
}

var Metrics *PrometheusMetrics

func InitMetrics() {
	Metrics = &PrometheusMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "myapp",              // Senior Notu 2: Namespace kullanımı karışıklığı önler.
				Subsystem: "ping_service",       // Alt sistem adı.
				Name:      "http_requests_total",
				Help:      "Number of requests received by status code.",
			},
			[]string{"path", "method", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "ping_service",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				// Senior Notu 3: Custom Buckets. SLA'inize göre ayarlayın.
				// Örneğin: 50ms, 100ms, 200ms, 500ms, 1s, 2s, 5s
				Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5},
			},
			[]string{"path", "method"},
		),
		// Response Size
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "ping_service",
				Name:      "http_response_size_bytes",
				Help:      "Size of HTTP responses in bytes.",
				// Byte bucketları: 1KB, 10KB, 100KB, 1MB...
				Buckets: prometheus.ExponentialBuckets(100, 10, 6),
			},
			[]string{"path", "method"},
		),
		InFlightRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "myapp",
				Subsystem: "ping_service",
				Name:      "http_requests_in_flight",
				Help:      "Current number of requests being processed.",
			},
			[]string{"path", "method"},
		),
		DependencyDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "myapp",
				Subsystem: "ping_service",
				Name:      "dependency_duration_seconds",
				Help:      "Duration of outgoing requests to other services.",
				Buckets:   []float64{0.05, 0.1, 0.5, 1, 2},
			},
			[]string{"target_service", "method"},
		),
	}

	prometheus.MustRegister(Metrics.RequestsTotal)
	prometheus.MustRegister(Metrics.RequestDuration)
	prometheus.MustRegister(Metrics.ResponseSize)
	prometheus.MustRegister(Metrics.InFlightRequests)
	prometheus.MustRegister(Metrics.DependencyDuration)
}

func PrometheusMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		// Senior Notu 4: Path Sanitization (High Cardinality Koruması)
		// c.Request().URL.Path kullanırsan "/users/123" gibi unique ID'ler metrikleri patlatır.
		// Echo'da c.Path() genellikle route pattern'i döner ("/users/:id").
		// Eğer path boşsa, "unknown" veya "root" demek daha güvenlidir.
		path := c.Path()
		if path == "" {
			path = "unknown_route" // Raw path yerine bunu tercih et.
		}
		method := c.Request().Method

		Metrics.InFlightRequests.WithLabelValues(path, method).Inc()
		defer Metrics.InFlightRequests.WithLabelValues(path, method).Dec()

		err := next(c)

		status := c.Response().Status
		// Echo hataları (404, 500 vb.) bazen error objesi olarak döner, status set edilmeyebilir.
		if err != nil {
			c.Error(err) // Echo'nun hata handler'ını tetikle
			status = c.Response().Status
		}

		duration := time.Since(start).Seconds()
		// Response size'ı alıyoruz
		respSize := float64(c.Response().Size)

		Metrics.RequestsTotal.WithLabelValues(path, method, strconv.Itoa(status)).Inc()
		Metrics.RequestDuration.WithLabelValues(path, method).Observe(duration)
		Metrics.ResponseSize.WithLabelValues(path, method).Observe(respSize) // Bytes grafiği için

		return err
	}
}

func MetricsHandler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.Handler())
}