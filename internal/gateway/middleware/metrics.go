package middleware

import (
	"fmt"
	"time"

	"github.com/Nzyazin/zadnik.store/internal/gateway/metrics"
	"github.com/gin-gonic/gin"
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		metrics.ActiveConnections.Inc()

		defer func() {
			metrics.ActiveConnections.Dec()

			duration := time.Since(start).Seconds()
			metrics.HTTPRequestDuration.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(duration)

			metrics.HTTPRequestsTotal.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
				fmt.Sprint(c.Writer.Status()),
			).Inc()
		}()

		c.Next()
	}
}