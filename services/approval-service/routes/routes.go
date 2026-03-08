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
