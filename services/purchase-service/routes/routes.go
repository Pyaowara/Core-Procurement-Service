package routes

import (
	"net/http"
	"strconv"
	"time"

	"github.com/core-procurement/purchase-service/handlers"
	"github.com/core-procurement/purchase-service/middleware"
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
	r.Use(metricsMiddleware("purchase-service"))
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "purchase-service",
		})
	})

	// PR operations - for Employee/Manager to create and manage PRs
	pr := r.Group("/pr")
	{
		pr.POST("", middleware.EmployeeRequired(), handlers.CreatePR)
		pr.GET("", middleware.AuthRequired(), handlers.GetPRList)
		pr.GET("/:id", middleware.AuthRequired(), handlers.GetPR)
		pr.PUT("/:id", middleware.EmployeeRequired(), handlers.UpdatePR)
		pr.POST("/:id/submit", middleware.EmployeeRequired(), handlers.SubmitPR)
		pr.DELETE("/:id", middleware.EmployeeRequired(), handlers.DeletePR)
	}

	// PO operations - for PurchaseOfficer to manage purchase orders
	po := r.Group("/po")
	{
		po.POST("", middleware.PurchaseOfficerRequired(), handlers.GeneratePO)
		po.GET("", middleware.AuthRequired(), handlers.GetPOList)
		po.GET("/:id", middleware.AuthRequired(), handlers.GetPO)
		po.PUT("/:id", middleware.PurchaseOfficerRequired(), handlers.ReceiveGoods)
		po.DELETE("/:id", middleware.PurchaseOfficerRequired(), handlers.DeletePO)
	}

	// Vendor management - for Admin users
	vendor := r.Group("/vendor")
	{
		vendor.POST("", middleware.AdminOnly(), handlers.CreateVendor)
		vendor.GET("", middleware.AuthRequired(), handlers.GetVendorList)
		vendor.GET("/:id", middleware.AuthRequired(), handlers.GetVendor)
		vendor.PUT("/:id", middleware.AdminOnly(), handlers.UpdateVendor)
		vendor.DELETE("/:id", middleware.AdminOnly(), handlers.DeleteVendor)
	}

	return r
}
