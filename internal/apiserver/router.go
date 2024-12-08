package apiserver

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewAPIServer initializes and runs a new API server on port 8080. It returns an error if the server fails to start.
func NewAPIServer(etcdClient *EtcdClient) error {
	router := setupRouter(etcdClient)

	err := router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

// setupRouter initializes and returns a new Gin Engine with predefined routes for health checks and API endpoints.
func setupRouter(etcdClient *EtcdClient) *gin.Engine {
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
		router.GET("/"+endpoint+"/:key", func(c *gin.Context) {
			key := c.Param("key")

			// Retrieve data from etcd
			value, err := etcdClient.GetData(key)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"value": value,
			})
		})

		// Write routes
		router.POST("/"+endpoint, func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			// Store data to etcd
			for key, value := range data {
				if err := etcdClient.PutData(key, value.(string)); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Data stored in etcd",
			})
		})
	}
	return router
}
