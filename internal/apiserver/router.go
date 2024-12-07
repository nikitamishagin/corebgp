package apiserver

import "github.com/gin-gonic/gin"

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
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		c.String(200, "ok")
	})

	router.GET("/announces/v1beta", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "announces v1beta endpoint",
		})
	})

	router.GET("/components/v1beta", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "components v1beta endpoint",
		})
	})

	return router
}
