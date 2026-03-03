package middleware

import (
	"net/http"
	"strings"

	"github.com/core-procurement/auth-identity-service/utils"
	"github.com/gin-gonic/gin"
)


func authenticate(c *gin.Context) bool {
	tokenString, err := c.Cookie("token")
	if err != nil || tokenString == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return false
	}

	claims, err := utils.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		c.Abort()
		return false
	}

	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	return true
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticate(c) {
			return
		}
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticate(c) {
			return
		}
		role, exists := c.Get("role")
		if !exists || role != "Admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			c.Abort()
			return
		}
		c.Next()
	}
}
