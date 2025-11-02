package observability

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware creates a Gin middleware that records HTTP metrics
func PrometheusMiddleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		
		// If no route match, use raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		metrics.RecordHTTPRequest(method, path, statusCode, duration)
	}
}
