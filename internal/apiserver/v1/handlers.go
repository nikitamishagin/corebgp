package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nikitamishagin/corebgp/internal/model"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
)

// Handler provides methods to manage and interact with a database using a DatabaseAdapter.
type Handler struct {
	DB model.DatabaseAdapter
}

// NewHandler initializes a new Handler instance with the provided DatabaseAdapter for interacting with the database.
func NewHandler(db model.DatabaseAdapter) *Handler {
	return &Handler{
		DB: db,
	}
}

// ListAnnouncements retrieves a list of all announcements by querying the database with a predefined prefix.
func (h *Handler) ListAnnouncements(c *gin.Context) {
	prefix := "v1/announcements/"

	announcementList, err := h.DB.List(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ListAnnouncementsResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.ListAnnouncementsResponse{
		Message: "Announcements retrieved successfully",
		Data:    announcementList,
	})
}

// GetAllAnnouncements retrieves all announcements from the database, deserializing them into structured data.
func (h *Handler) GetAllAnnouncements(c *gin.Context) {
	prefix := "v1/announcements/"

	data, err := h.DB.GetObjects(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GetAnnouncementsResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	announcements := make([]model.Announcement, 0, len(data))
	for _, value := range data {
		var announcement model.Announcement
		err = json.Unmarshal([]byte(value), &announcement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.GetAnnouncementsResponse{
				Message: "failed to unmarshal announcement",
				Data:    nil,
			})
			return
		}
		announcements = append(announcements, announcement)
	}

	c.JSON(http.StatusOK, model.GetAnnouncementsResponse{
		Message: "Announcements retrieved successfully",
		Data:    announcements,
	})
}

// ListAnnouncementsByProject retrieves a list of announcements for a specified project by querying the database.
func (h *Handler) ListAnnouncementsByProject(c *gin.Context) {
	// Extract params from path
	project := c.Param("project")
	prefix := "v1/announcements/" + project + "/"

	announcementList, err := h.DB.List(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ListAnnouncementsResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.ListAnnouncementsResponse{
		Message: "Announcements retrieved successfully",
		Data:    announcementList,
	})
}

// GetAllAnnouncementsByProject retrieves a list of announcements for a specified project by querying the database.
func (h *Handler) GetAllAnnouncementsByProject(c *gin.Context) {
	// Extract params from path
	project := c.Param("project")
	prefix := "v1/announcements/" + project + "/"

	data, err := h.DB.GetObjects(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.GetAnnouncementsResponse{
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	announcements := make([]model.Announcement, 0, len(data))
	for _, value := range data {
		var announcement model.Announcement
		err = json.Unmarshal([]byte(value), &announcement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.GetAnnouncementsResponse{
				Message: "failed to unmarshal announcement",
				Data:    nil,
			})
			return
		}
		announcements = append(announcements, announcement)
	}

	c.JSON(http.StatusOK, model.GetAnnouncementsResponse{
		Message: "Announcements retrieved successfully",
		Data:    announcements,
	})
}

// GetAnnouncement retrieves a specific announcement based on the project and name parameters from the database.
func (h *Handler) GetAnnouncement(c *gin.Context) {
	// Extract params from path
	project := c.Param("project")
	name := c.Param("name")

	prefix := "v1/announcements/" + project + "/" + name

	data, err := h.DB.Get(prefix)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.AnnouncementResponse{
			Message: "announcement not found",
			Data:    model.Announcement{},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
		return
	}

	var announcement model.Announcement
	err = json.Unmarshal([]byte(data), &announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: "failed to unmarshal announcement",
			Data:    model.Announcement{},
		})
		return
	}

	c.JSON(http.StatusOK, model.AnnouncementResponse{
		Message: "Announcement retrieved successfully",
		Data:    announcement,
	})
}

// PostAnnouncement creates a new announcement by accepting JSON data, validating it, and storing it in the database.
func (h *Handler) PostAnnouncement(c *gin.Context) {
	var announcement model.Announcement
	if err := c.ShouldBindJSON(&announcement); err != nil {
		c.JSON(http.StatusBadRequest, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
		return
	}

	prefix := "v1/announcements/" + announcement.Meta.Project + "/" + announcement.Meta.Name
	_, err := h.DB.Get(prefix)
	if err == nil {
		c.JSON(http.StatusConflict, model.AnnouncementResponse{
			Message: "announcement already exists",
			Data:    model.Announcement{},
		})
		return
	}
	if err.Error() != "key not found" {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: fmt.Errorf("failed to check announcement existence: %w", err).Error(),
			Data:    model.Announcement{},
		})
		return
	}

	data, err := json.Marshal(announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
		return
	}

	err = h.DB.Put(prefix, string(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: fmt.Errorf("failed to write announcement: %w", err).Error(),
			Data:    model.Announcement{},
		})
		return
	}

	c.JSON(http.StatusCreated, model.AnnouncementResponse{
		Message: "Announcement created successfully",
		Data:    announcement,
	})
}

// PatchAnnouncements updates an existing announcement in the database based on provided JSON data.
func (h *Handler) PatchAnnouncements(c *gin.Context) {
	var announcement model.Announcement
	if err := c.ShouldBindJSON(&announcement); err != nil {
		c.JSON(http.StatusBadRequest, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
		return
	}

	prefix := "v1/announcements/" + announcement.Meta.Project + "/" + announcement.Meta.Name
	_, err := h.DB.Get(prefix)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.AnnouncementResponse{
			Message: "announcement not found",
			Data:    model.Announcement{},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
		return
	}

	data, err := json.Marshal(announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: err.Error(),
			Data:    model.Announcement{},
		})
	}

	err = h.DB.Put(prefix, string(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: fmt.Errorf("failed to patch announcement: %w", err).Error(),
			Data:    model.Announcement{},
		})
		return
	}

	c.JSON(http.StatusOK, model.AnnouncementResponse{
		Message: "Announcement patched successfully",
		Data:    announcement,
	})
}

// Declare WebSocket upgrader object
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any client
	},
}

// WatchAnnouncements establishes a WebSocket connection and streams announcements from the database to the client.
func (h *Handler) WatchAnnouncements(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.WatchEvent{
			Message: fmt.Errorf("failed to enseblish WebSocket connection: %w", err).Error(),
			Data:    model.Announcement{},
		})
		return
	}
	defer conn.Close()

	// Create a channel to stop the Watch
	stopChan := make(chan struct{})

	// Start watching keys with the prefix "/v1/announcements/"
	eventsChan, err := h.DB.Watch("v1/announcements/", stopChan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.WatchEvent{
			Message: fmt.Errorf("failed to start watching: %w", err).Error(),
			Data:    model.Announcement{},
		})
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
	for watchResponse := range eventsChan {
		for i := range watchResponse.Events {
			var watchEvent model.WatchEvent

			switch watchResponse.Events[i].Type {
			case clientv3.EventTypePut:
				if watchResponse.Events[i].IsCreate() {
					watchEvent.Type = model.EventAdded
				} else {
					watchEvent.Type = model.EventUpdated
				}

				err := json.Unmarshal(watchResponse.Events[i].Kv.Value, &watchEvent.Data)
				if err != nil {
					fmt.Printf("failed to unmarshal announcement: %v\n", err)
					continue
				}
			case clientv3.EventTypeDelete:
				watchEvent.Type = model.EventDeleted

				if watchResponse.Events[i].PrevKv != nil {
					err := json.Unmarshal(watchResponse.Events[i].PrevKv.Value, &watchEvent.Data)
					if err != nil {
						fmt.Printf("failed to unmarshal announcement: %v\n", err)
						continue
					}
				}
			}

			// Send the watchEvent to the client via WebSocket
			if err := conn.WriteJSON(watchEvent); err != nil {
				return
			}
		}
	}
}

// DeleteAnnouncement deletes a specific announcement from the database using the provided project and name parameters.
func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	// Extract params from path
	project := c.Param("project")
	name := c.Param("name")

	prefix := "v1/announcements/" + project + "/" + name
	_, err := h.DB.Get(prefix)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.AnnouncementResponse{
			Message: "announcement not found",
			Data:    model.Announcement{},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: fmt.Errorf("failed to check announcement existence: %w", err).Error(),
			Data:    model.Announcement{},
		})
	}

	data, err := h.DB.Delete(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: fmt.Errorf("failed to delete announcement: %w", err).Error(),
			Data:    model.Announcement{},
		})
		return
	}

	var announcement model.Announcement
	err = json.Unmarshal([]byte(data), &announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.AnnouncementResponse{
			Message: "Announcement deleted successfully, but failed to unmarshal announcement.",
			Data:    model.Announcement{},
		})
		return
	}

	c.JSON(http.StatusOK, model.AnnouncementResponse{
		Message: "Announcement deleted successfully",
		Data:    announcement,
	})
}
