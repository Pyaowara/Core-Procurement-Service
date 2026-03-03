package routes

import (
	"net/http"

	"github.com/core-procurement/auth-identity-service/handlers"
	"github.com/core-procurement/auth-identity-service/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "auth-identity-service",
		})
	})

	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
		auth.POST("/logout", handlers.Logout)
		auth.GET("/me", middleware.AuthRequired(), handlers.Me)
	}

	users := r.Group("/users")
	users.Use(middleware.AuthRequired())
	{
		users.GET("", middleware.AdminRequired(), handlers.GetAllUsers)
		users.GET("/:id", middleware.AdminRequired(), handlers.GetUser)
		users.PUT("/:id", handlers.UpdateUser)
		users.DELETE("/:id", middleware.AdminRequired(), handlers.DeleteUser)
		users.PATCH("/:id/role", middleware.AdminRequired(), handlers.UpdateRole)
	}

	return r
}
