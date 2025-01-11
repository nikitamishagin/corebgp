package apiserver

import (
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/spf13/cobra"
)

// RootCmd initializes and returns the root command for the CoreBGP API server application.
func RootCmd() *cobra.Command {
	var config model.APIConfig
	var cmd = &cobra.Command{
		Use:   "apiserver",
		Short: "CoreBGP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the database adapter
			databaseAdapter, err := initializeDatabaseAdapter(&config)
			if err != nil {
				return err
			}
			defer databaseAdapter.Close()

			// Start the API server
			if err := NewAPIServer(databaseAdapter); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config.DBType, "db-type", "etcd", "Database type")
	cmd.Flags().StringSliceVar(&config.DBEndpoints, "endpoints", []string{"http://localhost:2379"}, "Comma separated list of database endpoints")
	cmd.Flags().StringVar(&config.Etcd.CACert, "etcd-ca", "", "Path to etcd CA certificate")
	cmd.Flags().StringVar(&config.Etcd.ClientCert, "etcd-cert", "", "Path to etcd client certificate")
	cmd.Flags().StringVar(&config.Etcd.ClientKey, "etcd-key", "", "Path to etcd client key")
	cmd.Flags().StringVar(&config.TLSCert, "tls-cert", "", "Path to TLS certificate")
	cmd.Flags().StringVar(&config.TLSKey, "tls-key", "", "Path to TLS key")
	cmd.Flags().StringVarP(&config.LogPath, "log-path", "l", "/var/log/corebgp/apiserver.log", "Path to log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}

// initializeDatabaseAdapter initializes the appropriate database adapter based on the config.DBType value
func initializeDatabaseAdapter(config *model.APIConfig) (model.DatabaseAdapter, error) {
	switch config.DBType {
	case "etcd":
		// Initialize Etcd adapter
		etcdClient, err := NewEtcdClient(config.DBEndpoints, config.Etcd.CACert, config.Etcd.ClientCert, config.Etcd.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize etcd adapter: %w", err)
		}
		return etcdClient, nil

	default:
		// Return an error if DBType is unknown
		return nil, fmt.Errorf("unsupported db type: %s", config.DBType)
	}
}

// TODO: Implement config validation

// validateEndpoints validates a list of endpoint URLs and ensures they have proper format, schema, host, and port.
// Returns a slice of sanitized endpoints or an error if validation fails.
//func validateEndpoints(endpoints []string) ([]string, error) {
//	result := make([]string, len(endpoints))
//
//	// Checking that all elements in a list are valid and parsing them
//	for i := range endpoints {
//		baseURL := strings.TrimSpace(endpoints[i])
//		if baseURL == "" {
//			return []string{}, fmt.Errorf("endpoint cannot be empty")
//		}
//
//		// Parse the URL to validate schema and host
//		parsedURL, err := url.Parse(baseURL)
//		if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
//			return nil, fmt.Errorf("endpoint must be in format proto://host:port, got: %s", baseURL)
//		}
//
//		// Split host and port
//		host, port, err := net.SplitHostPort(parsedURL.Host)
//		if err != nil {
//			return nil, fmt.Errorf("endpoint must be in format proto://host:port, got: %s", baseURL)
//		}
//
//		// Validate port
//		portNum, err := strconv.ParseUint(port, 10, 32)
//		if err != nil || portNum < 1 || portNum > 65535 {
//			return nil, fmt.Errorf("port must be a number between 1 and 65535, got: %s", port)
//		}
//
//		result[i] = parsedURL.Host
//	}
//
//	return result, nil
//}
