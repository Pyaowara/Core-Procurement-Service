package handlers

import (
	"net/http"
	"github.com/core-procurement/inventory-service/config"
	"github.com/core-procurement/inventory-service/models"
	"github.com/gin-gonic/gin"
)

type ViewInventory struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	UnitPrice   float64 `json:"unit_price"`
}

func CreateInventory(c *gin.Context) {
	var input models.Inventory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := config.GetDB()
	if err := db.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func GetInventories(c *gin.Context) {
	var inventories []models.Inventory
	db := config.GetDB()
	if err := db.Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventories)
}

func GetInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	db := config.GetDB()
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func UpdateInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	db := config.GetDB()
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	var input models.Inventory
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Model(&inventory).Updates(input)
	c.JSON(http.StatusOK, inventory)
}

func DeleteInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	db := config.GetDB()
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	db.Delete(&inventory)
	c.JSON(http.StatusOK, gin.H{"message": "Inventory deleted"})
}

func GetInventoryList(c *gin.Context) {
	var inventories []ViewInventory
	db := config.GetDB()
	if err := db.Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventories)
}