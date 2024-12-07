package apiserver

import "github.com/gin-gonic/gin"

func NewAPIServer() error {
	router := setupRouter()

	err := router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

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
