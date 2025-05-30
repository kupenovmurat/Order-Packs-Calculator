package main

import (
	"log"
	"net/http"
	"os"
	"pack-calculator/internal/handlers"
	"pack-calculator/internal/models"
	"pack-calculator/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	config := models.NewPackConfiguration()
	packService := service.NewPackCalculatorService(config)
	packHandler := handlers.NewPackHandler(packService)
	router := setupRouter(packHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting pack calculator server on port %s", port)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /health                 - Health check")
	log.Printf("  GET  /                       - Web UI")
	log.Printf("  GET  /api/calculate?items=N  - Calculate packs (query param)")
	log.Printf("  POST /api/calculate          - Calculate packs (JSON body)")
	log.Printf("  GET  /api/pack-sizes         - Get current pack sizes")
	log.Printf("  PUT  /api/pack-sizes         - Update pack sizes")

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(packHandler *handlers.PackHandler) *gin.Engine {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	router.GET("/health", packHandler.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/calculate", packHandler.CalculatePacksQuery)
		api.POST("/calculate", packHandler.CalculatePacks)
		api.GET("/pack-sizes", packHandler.GetPackSizes)
		api.PUT("/pack-sizes", packHandler.UpdatePackSizes)
	}

	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Pack Calculator",
		})
	})

	return router
}
