package handlers

import (
	"net/http"

	"github.com/core-procurement/purchase-service/config"
	"github.com/core-procurement/purchase-service/models"
	"github.com/gin-gonic/gin"
)

type CreateVendorRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
}

type UpdateVendorRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
}

// CreateVendor creates a new vendor
func CreateVendor(c *gin.Context) {
	var req CreateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vendor := models.Vendor{
		Name:    req.Name,
		Address: req.Address,
		TaxID:   req.TaxID,
	}

	if err := config.DB.Create(&vendor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create vendor"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Vendor created successfully",
		"data":    vendor,
	})
}

// UpdateVendor updates an existing vendor
func UpdateVendor(c *gin.Context) {
	vendorID := c.Param("id")

	var req UpdateVendorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var vendor models.Vendor
	if err := config.DB.First(&vendor, vendorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
		return
	}

	if req.Name != "" {
		vendor.Name = req.Name
	}
	if req.Address != "" {
		vendor.Address = req.Address
	}
	if req.TaxID != "" {
		vendor.TaxID = req.TaxID
	}

	config.DB.Save(&vendor)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vendor updated successfully",
		"data":    vendor,
	})
}

// GetVendor retrieves a vendor by ID
func GetVendor(c *gin.Context) {
	vendorID := c.Param("id")

	var vendor models.Vendor
	if err := config.DB.First(&vendor, vendorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
		return
	}

	c.JSON(http.StatusOK, vendor)
}

// GetVendorList retrieves all vendors
func GetVendorList(c *gin.Context) {
	var vendors []models.Vendor

	if err := config.DB.Find(&vendors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve vendors"})
		return
	}

	c.JSON(http.StatusOK, vendors)
}

// DeleteVendor soft deletes a vendor
func DeleteVendor(c *gin.Context) {
	vendorID := c.Param("id")

	var vendor models.Vendor
	if err := config.DB.First(&vendor, vendorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
		return
	}

	config.DB.Delete(&vendor)

	c.JSON(http.StatusOK, gin.H{"message": "Vendor deleted successfully"})
}
