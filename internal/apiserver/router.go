package apiserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/nikitamishagin/corebgp/internal/model"
	"net/http"
)

// NewAPIServer initializes and runs a new API server on port 8080. It returns an error if the server fails to start.
func NewAPIServer(databaseAdapter model.DatabaseAdapter) error {
	router := setupRouter(databaseAdapter)

	err := router.Run(":8080")
	if err != nil {
		return err
	}

	return nil
}

// setupRouter initializes and returns a new Gin Engine with predefined routes for health checks and API endpoints.
func setupRouter(db model.DatabaseAdapter) *gin.Engine {
	router := gin.Default()

	router.GET("/healthz", func(c *gin.Context) {
		// Check connection to etcd
		if err := db.HealthCheck(); err != nil {
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
		value, err := db.Get(key)
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

		key := "v1/announces/" + data.Meta.Project + "/" + data.Meta.Name
		_, err := db.Get(key)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "announce already exists"})
			return
		}
		if err != nil && err.Error() != "Key not found" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check announce existence"})
			return
		}

		value, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		err = db.Put("v1/announces/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Data written successfully"})
	})

	v1.PATCH("/announces/:project/:name", func(c *gin.Context) {
		var data model.Announce
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := "v1/announces/" + data.Meta.Project + "/" + data.Meta.Name
		_, err := db.Get(key)
		if err != nil && err.Error() == "Key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "announce not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check announce existence"})
			return
		}

		value, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		err = db.Put("v1/announces/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Announce updated successfully"})
	})
	return router
}
