package handlers

import (
	"net/http"

	"github.com/core-procurement/auth-identity-service/config"
	"github.com/core-procurement/auth-identity-service/models"
	"github.com/core-procurement/auth-identity-service/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Role      string `json:"role" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		Username:  input.Username,
		Password:  string(hashedPassword),
		Role:      input.Role,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
		},
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.SetCookie("token", token, 60*60*24, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{
		"message": "login successful",
		"token":   token,
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"role":       user.Role,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"email":      user.Email,
		},
	})
}

func Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func Me(c *gin.Context) {
	userID, _ := c.Get("user_id")
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
		"role":     role,
	})
}
