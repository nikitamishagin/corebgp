package apiserver

import "github.com/gin-gonic/gin"

func NewAPIServer() error {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	err := router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}
