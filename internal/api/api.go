package api

import (
	"github.com/nikitamishagin/corebgp/internal/etcd"
	"github.com/spf13/cobra"
	"log"

	"github.com/gin-gonic/gin"
)

type APIServer struct {
	client *etcd.Client
	router *gin.Engine
}

type Config struct {
	TLSCert       string
	TLSKey        string
	GobgpEndpoint string
	LogPath       string
	Verbose       int8
}

var (
	etcdEndpoints string
	etcdCert      string
	etcdKey       string
	tlsCert       string
	tlsKey        string
	gobgpInstance string
	logPath       string
	verbose       int8
)

func Run() error {
	var rootCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "API Server is a RESTful server for interacting with etcd",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := etcd.NewClient(etcd.Config{
				Endpoints: []string{etcdEndpoints},
				Cert:      etcdCert,
				Key:       etcdKey,
			})
			if err != nil {
				log.Fatalf("Failed to connect to etcd: %v", err)
			}

			serverConfig := Config{
				TLSCert:       tlsCert,
				TLSKey:        tlsKey,
				GobgpEndpoint: gobgpInstance,
				LogPath:       logPath,
				Verbose:       verbose,
			}
			server := NewAPIServer(client, serverConfig)
			server.Start()
		},
	}

	rootCmd.Flags().StringVar(&etcdEndpoints, "etcd-endpoints", "http://localhost:2379", "Comma separated list of etcd endpoints")
	rootCmd.Flags().StringVar(&etcdCert, "etcd-cert", "", "Path to etcd client certificate")
	rootCmd.Flags().StringVar(&etcdKey, "etcd-key", "", "Path to etcd client key")
	rootCmd.Flags().StringVar(&tlsCert, "tls-cert", "", "Path to TLS certificate")
	rootCmd.Flags().StringVar(&tlsKey, "tls-key", "", "Path to TLS key")
	rootCmd.Flags().StringVar(&gobgpInstance, "gobgp-instance", "http://localhost:50051", "Endpoint of gobgp instance")
	rootCmd.Flags().StringVarP(&logPath, "log-path", "l", "/var/log/corebgp/api.log", "Path to log file")
	rootCmd.Flags().Int8VarP(&verbose, "verbose", "v", 0, "Verbosity level")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}

	return nil
}

func NewAPIServer(client *etcd.Client, serverConfig Config) *APIServer {
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
