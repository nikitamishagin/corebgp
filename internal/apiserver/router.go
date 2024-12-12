package apiserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nikitamishagin/corebgp/internal/model"
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
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		// Check connection to etcd
		if err := etcdClient.CheckHealth(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "ok")
	})

	v1 := router.Group("/v1")

	v1.GET("/announces/:project/:name", func(c *gin.Context) {
		// Extract params from path
		project := c.Param("project")
		name := c.Param("name")

		// Create key for etcd data
		key := "v1/announces/" + project + "/" + name

		// Retrieve data from etcd
		value, err := etcdClient.GetData(key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var announce model.Announce
		err = json.Unmarshal([]byte(value), &announce)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal announce"})
			return
		}

		c.JSON(http.StatusOK, announce)
	})

	// Write routes
	v1.POST("/announces/", func(c *gin.Context) {
		var data model.Announce
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		value, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		err = etcdClient.PutData("v1/announces/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Data written successfully"})
	})

	return router
}
