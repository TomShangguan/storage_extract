package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// StartGinServer starts the Gin web server
func StartGinServer(port string) error {
	// Create Gin router with default middleware (logger, recovery)
	r := gin.Default()

	// Setup CORS middleware to allow frontend requests
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup API routes with Gin handlers
	setupGinAPIHandlers(r)

	// Setup static file serving for frontend
	setupGinStaticFileServer(r)

	// Start HTTP server
	fmt.Printf("Starting Ethereum Storage Visualizer (Gin) on port %s...\n", port)
	return r.Run(":" + port)
}

// setupGinAPIHandlers registers API endpoint handlers with Gin
func setupGinAPIHandlers(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/account/create", ginHandleCreateAccount)
		api.POST("/account/get", ginHandleGetAccount)
		api.POST("/storage/update", ginHandleUpdateStorage)
		api.POST("/storage/get", ginHandleGetValue)
		api.POST("/proof", ginHandleProof)
	}
}

// setupGinStaticFileServer configures static file serving with Gin
func setupGinStaticFileServer(r *gin.Engine) {
	// First check if front directory exists
	frontDir := "./front"
	if _, err := os.Stat(frontDir); os.IsNotExist(err) {
		// If front directory doesn't exist, try frontend directory
		frontDir = "./frontend"
		if _, err := os.Stat(frontDir); os.IsNotExist(err) {
			// If neither directory exists, print warning
			fmt.Println("WARNING: Neither 'front' nor 'frontend' directory exists!")
			fmt.Println("Please create either 'front' or 'frontend' directory with HTML/CSS/JS files.")
			return
		}
	}

	// Print the frontend directory being used
	absPath, _ := filepath.Abs(frontDir)
	fmt.Printf("Serving frontend files from: %s\n", absPath)

	// Set up static file server with Gin
	r.Static("/css", frontDir+"/css")
	r.Static("/js", frontDir+"/js")
	r.StaticFile("/", frontDir+"/index.html")
	r.StaticFile("/index.html", frontDir+"/index.html")
}
