package apiserver

import (
	"github.com/nikitamishagin/corebgp/internal/etcd"
	"log"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	client *etcd.Client
	router *gin.Engine
}

func NewAPIServer(client *etcd.Client) *APIServer {
	server := &APIServer{
		client: client,
		router: gin.Default(),
	}
	server.setupRoutes()
	return server
}

func (s *APIServer) setupRoutes() {
	s.router.GET("/", s.handleRoot)
	// Вы можете добавлять другие маршруты ниже
}

func (s *APIServer) Start() {
	log.Println("Starting API server on :8080")
	if err := s.router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *APIServer) handleRoot(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome to the API Server"})
}
