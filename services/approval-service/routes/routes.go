package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/core-procurement/approval-service/handlers"
	"github.com/core-procurement/approval-service/middleware"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"service", "method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func metricsMiddleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(service, c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(service, c.Request.Method, path, status).Observe(time.Since(start).Seconds())
	}
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Use(metricsMiddleware("approval-service"))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "approval-service",
		})
	})

	// Approval endpoints - at root level, no /approvals prefix
	r.GET("/:entity_type/:entity_id", middleware.AuthRequired(), handlers.GetApproval)
	r.POST("/:id/approve", middleware.AuthRequired(), handlers.ApproveStep)
	r.POST("/:id/reject", middleware.AuthRequired(), handlers.RejectStep)

	// Workflow ID based endpoints
	r.GET("/workflows/:workflow_id", middleware.AuthRequired(), handlers.GetApprovalByWorkflow)
	r.POST("/workflows/:workflow_id/approve", middleware.AuthRequired(), handlers.ApproveStepByWorkflow)
	r.POST("/workflows/:workflow_id/reject", middleware.AuthRequired(), handlers.RejectStepByWorkflow)

	return r
}
