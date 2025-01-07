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

// setupRouter initializes and returns a new Gin Engine with predefined routes for health checks and API endpoints.
func setupRouter(db model.DatabaseAdapter) *gin.Engine {
	router := gin.Default()

	v1Handler := v1.NewHandler(db)

	router.GET("/healthz", func(c *gin.Context) {
		// Check connection to etcd
		if err := db.HealthCheck(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.String(http.StatusOK, "ok")
	})

	v1 := router.Group("/v1")

	v1.GET("/announcements/", v1Handler.ListAnnouncements)

	v1.GET("/announcements/all", v1Handler.GetAnnouncements)

	v1.GET("/announcements/:project/", v1Handler.ListAnnouncementsByProject)

	v1.GET("/announcements/:project/all", v1Handler.GetAnnouncementsByProject)

	v1.GET("/announcements/:project/:name", v1Handler.GetAnnouncement)

	// Write routes
	v1.POST("/announcements/", v1Handler.PostAnnouncement)

	v1.PATCH("/announcements/", v1Handler.PatchAnnouncements)

	// Route for watching announcements
	v1.GET("/watch/announcements/", v1Handler.WatchAnnouncements)

	v1.DELETE("/announcements/:project/:name", v1Handler.DeleteAnnouncement)

	return router
}
