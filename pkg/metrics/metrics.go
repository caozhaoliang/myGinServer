package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics 结构体封装所有监控指标
type Metrics struct {
	// HTTP 请求计数器
	httpRequestsTotal *prometheus.CounterVec

	// HTTP 请求耗时直方图
	httpRequestDuration *prometheus.HistogramVec

	// 正在处理的请求数
	httpRequestsInFlight prometheus.Gauge

	// 业务相关指标
	activeUsers     prometheus.Gauge
	ordersProcessed *prometheus.CounterVec
	cacheHits       prometheus.Counter
	cacheMisses     prometheus.Counter
}

// NewMetrics 创建新的监控指标实例
func NewMetrics() *Metrics {
	return &Metrics{
		httpRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),

		httpRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		httpRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		activeUsers: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_users",
				Help: "Current number of active users",
			},
		),

		ordersProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "orders_processed_total",
				Help: "Total number of processed orders",
			},
			[]string{"status"}, // success, failed
		),

		cacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_hits_total",
				Help: "Total number of cache hits",
			},
		),

		cacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "cache_misses_total",
				Help: "Total number of cache misses",
			},
		),
	}
}

// GinMiddleware 返回 Gin 中间件用于收集 HTTP 指标
func (m *Metrics) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过 metrics 端点的监控
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		m.httpRequestsInFlight.Inc()

		c.Next()

		// 请求处理完成后记录指标
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		m.httpRequestsInFlight.Dec()
		m.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()

		m.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// IncActiveUsers 业务指标操作方法
func (m *Metrics) IncActiveUsers() {
	m.activeUsers.Inc()
}

func (m *Metrics) DecActiveUsers() {
	m.activeUsers.Dec()
}

func (m *Metrics) IncOrdersProcessed(status string) {
	m.ordersProcessed.WithLabelValues(status).Inc()
}

func (m *Metrics) IncCacheHits() {
	m.cacheHits.Inc()
}

func (m *Metrics) IncCacheMisses() {
	m.cacheMisses.Inc()
}
