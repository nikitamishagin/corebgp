package apiserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewAPIServer initializes and runs a new API server on port 8080. It returns an error if the server fails to start.
func NewAPIServer() error {
	router := setupRouter()

	err := router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

// setupRouter initializes and returns a new Gin Engine with predefined routes for health checks and API endpoints.
func setupRouter() *gin.Engine {
	endpoints := []string{
		"announces/v1beta",
		"components/v1beta",
	}

	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	for _, endpoint := range endpoints {
		// Read routes
		router.GET("/"+endpoint, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": endpoint + " endpoint",
			})
		})

		// Write routes
		router.POST("/"+endpoint, func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Test data received for " + endpoint,
				"data":    data,
			})
		})
	}
	return router
}
