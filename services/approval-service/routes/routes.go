package routes

import (
	"net/http"

	"github.com/core-procurement/approval-service/handlers"
	"github.com/core-procurement/approval-service/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "approval-service",
		})
	})

	// Approval endpoints
	approvals := r.Group("/approvals")
	approvals.Use(middleware.AuthRequired())
	{
		approvals.POST("", handlers.CreateApproval)
		approvals.GET("/:entity_type/:entity_id", handlers.GetApproval)
		approvals.POST("/:id/approve", handlers.ApproveStep)
		approvals.POST("/:id/reject", handlers.RejectStep)

		// Workflow ID based endpoints (decoupled from approval ID)
		approvals.GET("/workflow/:workflow_id", handlers.GetApprovalByWorkflow)
		approvals.POST("/workflow/:workflow_id/approve", handlers.ApproveStepByWorkflow)
		approvals.POST("/workflow/:workflow_id/reject", handlers.RejectStepByWorkflow)
	}

	return r
}
