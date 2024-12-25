package apiserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

	v1.GET("/announcements/", func(c *gin.Context) {
		prefix := "v1/announcements/"

		resp, err := db.List(prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"announcements": resp})
	})

	v1.GET("/announcements/all", func(c *gin.Context) {
		prefix := "v1/announcements/"

		resp, err := db.GetObjects(prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		announcementList := make([]model.Announcement, 0, len(resp))
		for _, value := range resp {
			var announcement model.Announcement
			err = json.Unmarshal([]byte(value), &announcement)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal announcement"})
				return
			}
			announcementList = append(announcementList, announcement)
		}

		c.JSON(http.StatusOK, announcementList)
	})

	v1.GET("/announcements/:project/", func(c *gin.Context) {
		project := c.Param("project")
		prefix := "v1/announcements/" + project + "/"

		resp, err := db.List(prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"announcements": resp})
	})

	v1.GET("/announcements/:project/all", func(c *gin.Context) {
		project := c.Param("project")
		prefix := "v1/announcements/" + project + "/"

		resp, err := db.GetObjects(prefix)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		announcementList := make([]model.Announcement, 0, len(resp))
		for _, value := range resp {
			var announcement model.Announcement
			err = json.Unmarshal([]byte(value), &announcement)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal announcement"})
				return
			}
			announcementList = append(announcementList, announcement)
		}

		c.JSON(http.StatusOK, announcementList)
	})

	v1.GET("/announcements/:project/:name", func(c *gin.Context) {
		// Extract params from path
		project := c.Param("project")
		name := c.Param("name")

		// Create key for etcd data
		key := "v1/announcements/" + project + "/" + name

		// Retrieve data from etcd
		value, err := db.Get(key)
		if err != nil && err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "announcement not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var announcement model.Announcement
		err = json.Unmarshal([]byte(value), &announcement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal announcement"})
			return
		}

		c.JSON(http.StatusOK, announcement)
	})

	// Write routes
	v1.POST("/announcements/", func(c *gin.Context) {
		var data model.Announcement
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := "v1/announcements/" + data.Meta.Project + "/" + data.Meta.Name
		_, err := db.Get(key)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "announcement already exists"})
			return
		}
		if err != nil && err.Error() != "key not found" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check announcement existence"})
			return
		}

		value, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = db.Put("v1/announcements/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "Announcement added successfully"})
	})

	v1.PATCH("/announcements/", func(c *gin.Context) {
		var data model.Announcement
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key := "v1/announcements/" + data.Meta.Project + "/" + data.Meta.Name
		_, err := db.Get(key)
		if err != nil && err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "announcement not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		value, err := json.Marshal(data)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		err = db.Put("v1/announcements/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Announcement updated successfully"})
	})

	// Declare WebSocket upgrader object
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any client
		},
	}

	// Route for watching announcements
	v1.GET("/watch/announcements/", func(c *gin.Context) {
		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to establish WebSocket connection"})
			return
		}
		defer conn.Close()

		// Create a channel to stop the Watch
		stopChan := make(chan struct{})

		// Start watching keys with the prefix "/v1/announcements/"
		eventsChan, err := db.Watch("v1/announcements/", stopChan)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start watching"})
			return
		}

		// Goroutine to read from WebSocket connection
		go func() {
			defer close(stopChan)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					// Stop work on read error (e.g., the client disconnected)
					return
				}
			}
		}()

		// Read changes from events and send them to the client
		for watchResp := range eventsChan {
			for _, event := range watchResp.Events {
				// Event structure to be transmitted
				watchEvent := gin.H{
					"type":  event.Type.String(), // Event type (PUT or DELETE)
					"key":   string(event.Kv.Key),
					"value": string(event.Kv.Value), // Value on PUT
				}

				// Send the event to the client via WebSocket
				if err := conn.WriteJSON(watchEvent); err != nil {
					return
				}
			}
		}
	})

	v1.DELETE("/announcements/:project/:name", func(c *gin.Context) {
		project := c.Param("project")
		name := c.Param("name")

		key := "v1/announcements/" + project + "/" + name
		_, err := db.Get(key)
		if err != nil && err.Error() == "key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "announcement not found"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check announcement existence"})
		}
		err = db.Delete(key)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Announcement deleted successfully"})
	})

	return router
}
