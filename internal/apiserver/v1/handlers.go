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

type Handler struct {
	DB model.DatabaseAdapter
}

func NewHandler(db model.DatabaseAdapter) *Handler {
	return &Handler{
		DB: db,
	}
}

func (h *Handler) ListAnnouncements(c *gin.Context) {
	prefix := "v1/announcements/"

	data, err := h.DB.List(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcements retrieved successfully",
		Data:    data,
	})
}

func (h *Handler) GetAnnouncements(c *gin.Context) {
	prefix := "v1/announcements/"

	data, err := h.DB.GetObjects(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	announcementList := make([]model.Announcement, 0, len(data))
	for _, value := range data {
		var announcement model.Announcement
		err = json.Unmarshal([]byte(value), &announcement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Status:  "error",
				Message: "failed to unmarshal announcement",
				Data:    nil,
			})
			return
		}
		announcementList = append(announcementList, announcement)
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcements retrieved successfully",
		Data:    announcementList,
	})
}

func (h *Handler) ListAnnouncementsByProject(c *gin.Context) {
	project := c.Param("project")
	prefix := "v1/announcements/" + project + "/"

	data, err := h.DB.List(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcements retrieved successfully",
		Data:    data,
	})
}

func (h *Handler) GetAnnouncementsByProject(c *gin.Context) {
	project := c.Param("project")
	prefix := "v1/announcements/" + project + "/"

	data, err := h.DB.GetObjects(prefix)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	announcementList := make([]model.Announcement, 0, len(data))
	for _, value := range data {
		var announcement model.Announcement
		err = json.Unmarshal([]byte(value), &announcement)
		if err != nil {
			c.JSON(http.StatusInternalServerError, model.APIResponse{
				Status:  "error",
				Message: "failed to unmarshal announcement",
				Data:    nil,
			})
			return
		}
		announcementList = append(announcementList, announcement)
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcements retrieved successfully",
		Data:    announcementList,
	})
}

func (h *Handler) GetAnnouncement(c *gin.Context) {
	// Extract params from path
	project := c.Param("project")
	name := c.Param("name")

	// Create key for etcd data
	key := "v1/announcements/" + project + "/" + name

	// Retrieve data from etcd
	value, err := h.DB.Get(key)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Status:  "error",
			Message: "announcement not found",
			Data:    nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	var announcement model.Announcement
	err = json.Unmarshal([]byte(value), &announcement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: "failed to unmarshal announcement",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcement retrieved successfully",
		Data:    announcement,
	})
}

func (h *Handler) PostAnnouncement(c *gin.Context) {
	var data model.Announcement
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	key := "v1/announcements/" + data.Meta.Project + "/" + data.Meta.Name
	_, err := h.DB.Get(key)
	if err == nil {
		c.JSON(http.StatusConflict, model.APIResponse{
			Status:  "error",
			Message: "announcement already exists",
			Data:    nil,
		})
		return
	}
	if err.Error() != "key not found" {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to check announcement existence: %w", err).Error(),
			Data:    nil,
		})
		return
	}

	value, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	err = h.DB.Put("v1/announcements/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to write announcement: %w", err).Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, model.APIResponse{
		Status:  "success",
		Message: "Announcement created successfully",
		Data: model.Event{
			Type:         model.EventAdded,
			Announcement: data,
		},
	})
}

func (h *Handler) PatchAnnouncements(c *gin.Context) {
	var data model.Announcement
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	key := "v1/announcements/" + data.Meta.Project + "/" + data.Meta.Name
	_, err := h.DB.Get(key)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Status:  "error",
			Message: "announcement not found",
			Data:    nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	value, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: err.Error(),
			Data:    nil,
		})
	}

	err = h.DB.Put("v1/announcements/"+data.Meta.Project+"/"+data.Meta.Name, string(value))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to patch announcement: %w", err).Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcement patched successfully",
		Data: model.Event{
			Type:         model.EventUpdated,
			Announcement: data,
		},
	})
}

// Declare WebSocket upgrader object
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any client
	},
}

func (h *Handler) WatchAnnouncements(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to enseblish WebSocket connection: %w", err).Error(),
			Data:    nil,
		})
		return
	}
	defer conn.Close()

	// Create a channel to stop the Watch
	stopChan := make(chan struct{})

	// Start watching keys with the prefix "/v1/announcements/"
	eventsChan, err := h.DB.Watch("v1/announcements/", stopChan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to start watching: %w", err).Error(),
			Data:    nil,
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
	for watchResp := range eventsChan {
		for _, watchEvent := range watchResp.Events {
			var eventResp model.Event

			switch watchEvent.Type {
			case clientv3.EventTypePut:
				if watchEvent.IsCreate() {
					eventResp.Type = model.EventAdded
				} else {
					eventResp.Type = model.EventUpdated
				}

				err := json.Unmarshal(watchEvent.Kv.Value, &eventResp.Announcement)
				if err != nil {
					fmt.Printf("failed to unmarshal announcement: %v\n", err)
					continue
				}
			case clientv3.EventTypeDelete:
				eventResp.Type = model.EventDeleted

				if watchEvent.PrevKv != nil {
					err := json.Unmarshal(watchEvent.PrevKv.Value, &eventResp.Announcement)
					if err != nil {
						fmt.Printf("failed to unmarshal announcement: %v\n", err)
						continue
					}
				}
			}

			// Send the eventResp to the client via WebSocket
			if err := conn.WriteJSON(eventResp); err != nil {
				return
			}
		}
	}
}

func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	project := c.Param("project")
	name := c.Param("name")

	key := "v1/announcements/" + project + "/" + name
	_, err := h.DB.Get(key)
	if err != nil && err.Error() == "key not found" {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Status:  "error",
			Message: "announcement not found",
			Data:    nil,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to check announcement existence: %w", err).Error(),
			Data:    nil,
		})
	}

	err = h.DB.Delete(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Status:  "error",
			Message: fmt.Errorf("failed to delete announcement: %w", err).Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Status:  "success",
		Message: "Announcement deleted successfully",
		Data: model.Event{
			Type:         model.EventDeleted,
			Announcement: model.Announcement{Meta: model.Meta{Project: project, Name: name}},
		},
	})
}
