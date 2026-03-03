package routes

import (
	"net/http"
	"github.com/core-procurement/inventory-service/middleware"
	"github.com/core-procurement/inventory-service/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

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
		inventory.GET("", middleware.AuthRequired(), handlers.GetInventoryList)
	}

	return r
}
