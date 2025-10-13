package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpReqs = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "http_requests_total", Help: "HTTP requests"},
	[]string{"route", "method", "status"},
)

func init() {
	prometheus.MustRegister(httpReqs)
}

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		httpReqs.WithLabelValues(route, c.Request.Method, strconv.Itoa(c.Writer.Status())).Inc()
	}
}

// Mounts /metrics
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) { h.ServeHTTP(c.Writer, c.Request) }
}
