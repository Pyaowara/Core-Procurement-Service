package routes

import (
	"net/http"

	"github.com/core-procurement/purchase-service/handlers"
	"github.com/core-procurement/purchase-service/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

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
		pr.POST("", middleware.EmployeeCanCreatePR(), handlers.CreatePR)
		pr.GET("", middleware.AuthRequired(), handlers.GetPRList)
		pr.GET("/:id", middleware.AuthRequired(), handlers.GetPR)
		pr.PUT("/:id", middleware.EmployeeCanCreatePR(), handlers.UpdatePR)
		pr.POST("/:id/submit", middleware.EmployeeCanCreatePR(), handlers.SubmitPR)
		pr.GET("/:id/snapshot", middleware.AuthRequired(), handlers.GetPRSnapshot)
		pr.DELETE("/:id", middleware.EmployeeCanCreatePR(), handlers.DeletePR)
	}

	// PO operations - for PurchaseOfficer to manage purchase orders
	po := r.Group("/po")
	{
		po.POST("", middleware.PurchaseOfficerCanCreatePO(), handlers.GeneratePO)
		po.GET("", middleware.AuthRequired(), handlers.GetPOList)
		po.GET("/:id", middleware.AuthRequired(), handlers.GetPO)
		po.PUT("/:id", middleware.PurchaseOfficerCanCreatePO(), handlers.UpdatePOStatus)
		po.DELETE("/:id", middleware.PurchaseOfficerCanCreatePO(), handlers.DeletePO)
		po.POST("/:id/receive", middleware.ManagerCanManageGoods(), handlers.ReceiveGoods)
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
