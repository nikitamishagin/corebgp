package apiserver

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/nikitamishagin/corebgp/internal/apiserver/v1"
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

// setupRouter initializes a Gin engine, sets up API routes and handlers, and configures health check and API endpoints.
func setupRouter(db model.DatabaseAdapter) *gin.Engine {
	router := gin.Default()

	v1Handler := v1.NewHandler(db)

	// Health check endpoint to verify if the service and database (etcd) are operational
	router.GET("/healthz", func(c *gin.Context) {
		// Check connection to etcd
		if err := db.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "ok")
	})

	// Group all v1 API endpoints under /v1
	v1 := router.Group("/v1")

	// Retrieve a paginated list of announcements
	v1.GET("/announcements/", v1Handler.ListAnnouncements)

	// Retrieve all announcements without pagination
	v1.GET("/announcements/all", v1Handler.GetAnnouncements)

	// Retrieve a paginated list of announcements for a specific project
	v1.GET("/announcements/:project/", v1Handler.ListAnnouncementsByProject)

	// Retrieve all announcements for a specific project without pagination
	v1.GET("/announcements/:project/all", v1Handler.GetAnnouncementsByProject)

	// Retrieve a specific announcement by project and name
	v1.GET("/announcements/:project/:name", v1Handler.GetAnnouncement)

	// Create a new announcement
	v1.POST("/announcements/", v1Handler.PostAnnouncement)

	// Update specific fields of announcements
	v1.PATCH("/announcements/", v1Handler.PatchAnnouncements)

	// Watch announcements for real-time updates
	v1.GET("/watch/announcements/", v1Handler.WatchAnnouncements)

	// Delete a specific announcement by project and name
	v1.DELETE("/announcements/:project/:name", v1Handler.DeleteAnnouncement)

	return router
}
