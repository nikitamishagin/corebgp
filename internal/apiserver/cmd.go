package apiserver

import (
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

// RootCmd initializes and returns the root command for the CoreBGP API server application.
func RootCmd() *cobra.Command {
	var (
		endpointsList string
		config        model.APIConfig
	)
	var cmd = &cobra.Command{
		Use:   "apiserver",
		Short: "CoreBGP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse endpoints from the provided CLI argument
			endpoints, err := parseEndpoints(endpointsList)
			if err != nil {
				return err
			}
			config.Endpoints = endpoints

			// Initialize the database adapter
			databaseAdapter, err := initializeDatabaseAdapter(&config)
			if err != nil {
				return err
			}

			// Start the API server
			if err := NewAPIServer(databaseAdapter); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config.DBType, "db-type", "etcd", "Database type")
	cmd.Flags().StringVar(&endpointsList, "endpoints", "http://localhost:2379", "Comma separated list of database endpoints")
	cmd.Flags().StringVar(&config.Etcd.CACert, "etcd-ca", "", "Path to etcd CA certificate")
	cmd.Flags().StringVar(&config.Etcd.ClientCert, "etcd-cert", "", "Path to etcd client certificate")
	cmd.Flags().StringVar(&config.Etcd.ClientKey, "etcd-key", "", "Path to etcd client key")
	cmd.Flags().StringVar(&config.TLSCert, "tls-cert", "", "Path to TLS certificate")
	cmd.Flags().StringVar(&config.TLSKey, "tls-key", "", "Path to TLS key")
	cmd.Flags().StringVar(&config.GoBGPInstance, "gobgp-instance", "http://localhost:50051", "Endpoint of GoBGP instance")
	cmd.Flags().StringVarP(&config.LogPath, "log-path", "l", "/var/log/corebgp/apiserver.log", "Path to log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}

// initializeDatabaseAdapter initializes the appropriate database adapter based on the config.DBType value
func initializeDatabaseAdapter(config *model.APIConfig) (model.DatabaseAdapter, error) {
	switch config.DBType {
	case "etcd":
		// Initialize Etcd adapter
		etcdClient, err := NewEtcdClient(config)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize etcd adapter: %w", err)
		}
		return etcdClient, nil

	default:
		// Return an error if DBType is unknown
		return nil, fmt.Errorf("unsupported db type: %s", config.DBType)
	}
}

// parseEndpoints parses a comma-separated list of endpoints, validates each, and returns a slice of formatted endpoints.
// Returns an error if any endpoint is invalid, empty, or if the port is not within the valid range (1-65535).
func parseEndpoints(endpoints string) ([]string, error) {
	if endpoints == "" {
		return []string{}, fmt.Errorf("etcd endpoint cannot be empty")
	}

	endpointsSlice := strings.Split(endpoints, ",")
	var result []string

	// Checking that all elements in a list are valid and parsing them
	for i := range endpointsSlice {
		baseURL := strings.TrimSpace(endpointsSlice[i])
		if baseURL == "" {
			return []string{}, fmt.Errorf("endpoint cannot be empty")
		}

		protocolAndHost := strings.Split(baseURL, "//")
		if len(protocolAndHost) != 2 {
			return []string{}, fmt.Errorf("endpoint must be in format proto://host:port")
		}

		hostAndPort := strings.Split(protocolAndHost[1], ":")
		if len(hostAndPort) != 2 {
			return []string{}, fmt.Errorf("endpoint must be in format proto://host:port")
		}

		port, err := strconv.ParseUint(hostAndPort[1], 10, 64)
		if err != nil {
			return []string{}, fmt.Errorf("port must be a number")
		}
		if port < 1 || port > 65535 {
			return []string{}, fmt.Errorf("port must be between 1 and 65535")
		}

		result = append(result, protocolAndHost[1])
	}

	return result, nil
}
