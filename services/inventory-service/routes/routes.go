package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/core-procurement/inventory-service/handlers"
	"github.com/core-procurement/inventory-service/middleware"
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
	r.Use(metricsMiddleware("inventory-service"))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "inventory-service",
		})
	})

	dep := r.Group("/dep")
	{
		dep.POST("/inventory/create", middleware.DepRequired(), handlers.CreateInventory)
		dep.GET("/inventory", middleware.DepRequired(), handlers.GetInventories)
		dep.GET("/inventory/:id", middleware.DepRequired(), handlers.GetInventory)
		dep.PATCH("/inventory/:id", middleware.DepRequired(), handlers.UpdateInventory)
		dep.DELETE("/inventory/:id", middleware.DepRequired(), handlers.DeleteInventory)
	}

	inventory := r.Group("/inventory")
	{
		inventory.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "ok",
				"service": "inventory-service",
			})
		})
		inventory.GET("", middleware.AuthRequired(), handlers.GetInventoryList)
	}

	return r
}
