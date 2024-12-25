package updater

import (
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/nikitamishagin/corebgp/pkg/client/v1"
	"github.com/spf13/cobra"
	"time"
)

// RootCmd initializes and returns the root command for the CoreBGP API server application.
func RootCmd() *cobra.Command {
	var config model.UpdaterConfig
	var cmd = &cobra.Command{
		Use:   "updater",
		Short: "CoreBGP update controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Initialize the new GoBGP client
			goBGPClient, err := NewGoBGPClient(&config)
			if err != nil {
				return err
			}
			defer goBGPClient.Close()

			// Initialize the CoreBGP API client
			apiClient := v1.NewAPIClient(&config.APIEndpoint, time.Second*5)

			// Check if CoreBGP API server is healthy
			err = apiClient.V1HealthCheck(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&config.APIEndpoint, "api-endpoint", "http://localhost:8080", "URL of the API server")
	cmd.Flags().StringVar(&config.GoBGPEndpoint, "gobgp-endpoint", "127.0.0.1:50051", "GoBGP gRPC endpoint")
	cmd.Flags().StringVar(&config.GoBGPCACert, "gobgp-ca-cert", "", "Path to CA certificate")
	cmd.Flags().StringVar(&config.GoBGPClientCert, "gobgp-client-cert", "", "Path to client certificate")
	cmd.Flags().StringVar(&config.GoBGPClientKey, "gobgp-client-key", "", "Path to client key")
	cmd.Flags().StringVar(&config.LogPath, "log-path", "/var/log/corebgp/updater.log", "Path to the log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}
