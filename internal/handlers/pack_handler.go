package handlers

import (
	"net/http"
	"pack-calculator/internal/models"
	"pack-calculator/internal/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PackHandler struct {
	service *service.PackCalculatorService
}

func NewPackHandler(service *service.PackCalculatorService) *PackHandler {
	return &PackHandler{
		service: service,
	}
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func (h *PackHandler) CalculatePacks(c *gin.Context) {
	var req models.CalculateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result *models.PackResult
	var err error

	if len(req.PackSizes) > 0 {
		result, err = h.service.CalculatePacksWithCustomSizes(req.Items, req.PackSizes)
	} else {
		result, err = h.service.CalculatePacks(req.Items)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Calculation failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PackHandler) CalculatePacksQuery(c *gin.Context) {
	itemsStr := c.Query("items")
	if itemsStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Missing required parameter",
			Message: "items parameter is required",
		})
		return
	}

	items, err := strconv.Atoi(itemsStr)
	if err != nil || items <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid parameter",
			Message: "items must be a positive integer",
		})
		return
	}

	result, err := h.service.CalculatePacks(items)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Calculation failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *PackHandler) GetPackSizes(c *gin.Context) {
	packSizes := h.service.GetPackSizes()
	c.JSON(http.StatusOK, models.PackConfiguration{
		PackSizes: packSizes,
	})
}

func (h *PackHandler) UpdatePackSizes(c *gin.Context) {
	var config models.PackConfiguration

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	if err := h.service.UpdatePackSizes(config.PackSizes); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Update failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Pack sizes updated successfully",
		"pack_sizes": config.PackSizes,
	})
}

func (h *PackHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "pack-calculator",
		"version": "1.0.0",
	})
}
