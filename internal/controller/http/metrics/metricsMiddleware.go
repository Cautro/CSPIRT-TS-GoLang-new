package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ActiveRequests.Inc()
		defer ActiveRequests.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		statusCode := c.Writer.Status()
		statusStr := strconv.Itoa(statusCode)

		RequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			statusStr,
		).Inc()

		RequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		if statusCode >= 400 {
			errorClass := "4xx"
			if statusCode >= 500 {
				errorClass = "5xx"
			}

			ErrorsTotal.WithLabelValues(
				c.Request.Method,
				path,
				statusStr,
				errorClass,
			).Inc()
		}
	}
}