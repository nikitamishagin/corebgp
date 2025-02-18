package updater

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/nikitamishagin/corebgp/pkg/client/v1"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

// RootCmd initializes and returns the root command for the CoreBGP Updater application.
func RootCmd() *cobra.Command {
	var config model.UpdaterConfig
	var cmd = &cobra.Command{
		Use:   "updater",
		Short: "CoreBGP update controller",
		RunE: func(cmd *cobra.Command, args []string) error {
			for {
				// Create a context with cancel function for managing the goroutines
				ctx, cancel := context.WithCancel(cmd.Context())

				// Initialize the new GoBGP client
				goBGPClient, err := NewGoBGPClient(&config.GoBGPEndpoint, &config.GoBGPCACert, &config.GoBGPClientCert, &config.GoBGPClientKey)
				if err != nil {
					fmt.Printf("Failed to connect to GoBGP: %v. Retrying...\n", err)
					cancel()
					time.Sleep(5 * time.Second) // Retry interval
					continue
				}

				// TODO: Implement GoBGP configuration checking

				// Initialize the CoreBGP API client
				apiClient := v1.NewAPIClient(&config.APIEndpoint, time.Second*5)
				if err := apiClient.HealthCheck(ctx); err != nil {
					fmt.Printf("Failed to connect to CoreBGP API: %v. Retrying...\n", err)
					time.Sleep(5 * time.Second)
					continue
				}

				// Periodically check connectivity while the program is running
				go monitorConnectivity(ctx, goBGPClient, apiClient, cancel)

				fmt.Println("Connections established. Starting synchronization...")

				// Create a channel to process routeUpdates
				routeUpdates := make(chan RouteUpdate, 100) // Buffered channel to handle updates

				go watchAnnouncements(ctx, cancel, apiClient, routeUpdates)

				// Create channels for routes
				apiRoutesChan := make(chan map[string]Route, 1)
				controllerRoutesChan := make(chan map[string]Route, 1)

				// Create a WaitGroup to manage goroutines
				var wg sync.WaitGroup

				wg.Add(1)
				go fetchAPIRoutes(ctx, &wg, apiClient, apiRoutesChan)

				wg.Add(1)
				go fetchControllerRoutes(ctx, &wg, goBGPClient, controllerRoutesChan)

				wg.Add(1)
				go synchronizeRoutes(ctx, &wg, apiRoutesChan, controllerRoutesChan, goBGPClient)

				// Wait for sync goroutines to finish
				wg.Wait()

				fmt.Println("Handling routes...")

				wg.Add(1)
				go routesHanding(ctx, &wg, goBGPClient, routeUpdates)
				wg.Wait()

				fmt.Println("Closing connections...")
			}
		},
	}

	cmd.Flags().StringVar(&config.APIEndpoint, "api-endpoint", "http://localhost:8080", "URL of the API server")
	cmd.Flags().StringVar(&config.GoBGPEndpoint, "gobgp-endpoint", "localhost:50051", "GoBGP gRPC endpoint")
	cmd.Flags().StringVar(&config.GoBGPCACert, "gobgp-ca-cert", "", "Path to CA certificate")
	cmd.Flags().StringVar(&config.GoBGPClientCert, "gobgp-client-cert", "", "Path to client certificate")
	cmd.Flags().StringVar(&config.GoBGPClientKey, "gobgp-client-key", "", "Path to client key")
	cmd.Flags().StringVar(&config.LogPath, "log-path", "/var/log/corebgp/updater.log", "Path to the log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}

func monitorConnectivity(ctx context.Context, goBGPClient *GoBGPClient, apiClient *v1.APIClient, cancel context.CancelFunc) {
	ticker := time.NewTicker(10 * time.Second) // Check interval
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled, stopping connectivity monitoring...")
			return
		case <-ticker.C:
			// Check GoBGP connectivity
			if _, err := goBGPClient.GetBGP(); err != nil {
				fmt.Printf("Lost connection to GoBGP: %v\n", err)
				cancel() // Trigger reconnection
				return
			}

			// Check CoreBGP API connectivity
			if err := apiClient.HealthCheck(ctx); err != nil {
				fmt.Printf("Lost connection to CoreBGP API: %v\n", err)
				cancel() // Trigger reconnection
				return
			}
		}
	}
}
