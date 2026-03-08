package handlers

import (
	"net/http"

	"github.com/core-procurement/inventory-service/config"
	"github.com/core-procurement/inventory-service/models"
	"github.com/gin-gonic/gin"
)

type ViewInventory struct {
	ID          uint   `json:"id"`
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

type inventoryInput struct {
	Sku         string `json:"sku"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

func CreateInventory(c *gin.Context) {
	var input inventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inventory := models.Inventory{
		Sku:         input.Sku,
		Name:        input.Name,
		Description: input.Description,
		Quantity:    input.Quantity,
	}

	db := config.DB
	if err := db.Create(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"inventory": inventory})
}

func GetInventories(c *gin.Context) {
	var inventories []models.Inventory
	db := config.DB
	if err := db.Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, inventories)
}

func GetInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	db := config.DB
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	c.JSON(http.StatusOK, inventory)
}

func UpdateInventory(c *gin.Context) {
	id := c.Param("id")
	var inventory models.Inventory
	db := config.DB
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	var input inventoryInput
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
	db := config.DB
	if err := db.First(&inventory, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Inventory not found"})
		return
	}
	db.Delete(&inventory)
	c.JSON(http.StatusOK, gin.H{"message": "Inventory deleted"})
}

func GetInventoryList(c *gin.Context) {
	var inventories = []ViewInventory{}
	var inventory []models.Inventory
	db := config.DB
	if err := db.Find(&inventory).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(inventory) > 0 {
		for _, inv := range inventory {
			inventories = append(inventories, ViewInventory{
				ID:          inv.ID,
				Sku:         inv.Sku,
				Name:        inv.Name,
				Description: inv.Description,
				Quantity:    inv.Quantity,
			})
		}
	}
	c.JSON(http.StatusOK, inventories)
}
