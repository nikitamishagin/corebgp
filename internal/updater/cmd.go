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
			// Create a context with cancel function for managing the goroutines
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			// Create channels for routes
			apiRoutesChan := make(chan []Route, 1)
			controllerRoutesChan := make(chan []Route, 1)

			for {
				// Initialize the new GoBGP client
				goBGPClient, err := NewGoBGPClient(&config.GoBGPEndpoint, &config.GoBGPCACert, &config.GoBGPClientCert, &config.GoBGPClientKey)
				if err != nil {
					fmt.Printf("Failed to connect to GoBGP: %v. Retrying...\n", err)
					time.Sleep(5 * time.Second) // Retry interval
					continue
				}
				defer goBGPClient.Close()

				// TODO: Implement GoBGP configuration checking
				_, err = goBGPClient.GetBGP()
				if err != nil {
					fmt.Printf("Failed to get GoBGP configuration: %v. Retrying...\n", err)
					time.Sleep(5 * time.Second) // Retry interval
					continue
				}

				// For debug ListPath output
				//routes, err := goBGPClient.ListPath([]string{"192.168.1.1/32"})
				//if err != nil {
				//	log.Fatalf("Failed to list paths: %v", err)
				//}

				//fmt.Println("Output from CMD:")
				//fmt.Printf("Routes: %v\n", routes)

				// Initialize the CoreBGP API client
				apiClient := v1.NewAPIClient(&config.APIEndpoint, time.Second*5)

				// Check if CoreBGP API server is healthy
				err = apiClient.HealthCheck(ctx)
				if err != nil {
					fmt.Printf("Failed to connect to CoreBGP API: %v. Retrying...\n", err)
					time.Sleep(5 * time.Second) // Retry interval
					continue
				}

				// Launch synchronization logic when both APIs are reachable
				fmt.Println("Connections established. Starting synchronization...")

				// Create a WaitGroup to manage goroutines
				var wg sync.WaitGroup

				// Create a channel to process events
				events := make(chan model.Event, 100) // Buffered channel to handle bursts of events
				defer close(events)

				// TODO: Improve events handling and move to separate function
				// Goroutine for watching announcements
				wg.Add(1) // Increment the WaitGroup counter
				go func(ctx context.Context, cancel context.CancelFunc) {
					defer wg.Done() // Decrement the WaitGroup counter when the goroutine ends

					fmt.Println("Starting to watch announcements...")
					err := apiClient.WatchAnnouncements(ctx, func(event model.Event) {
						// Push each incoming event into the channel
						events <- event
					})
					if err != nil {
						fmt.Printf("Error while watching announcements: %v\n", err)
						cancel() // Cancel the context in case of an error
					}
				}(ctx, cancel) // Pass both context and cancel as arguments

				// Goroutine for processing events from the channel
				wg.Add(1) // Increment the WaitGroup counter
				go func() {
					defer wg.Done() // Ensure the WaitGroup counter is decremented after processing ends
					for event := range events {
						// Handle each event in a separate goroutine
						go func(ev model.Event) {
							if err := handleAnnouncementEvent(goBGPClient, &ev); err != nil {
								fmt.Printf("Failed to process event: %v\n", err)
							}
						}(event)
					}
				}()

				wg.Add(1)
				go fetchAPIRoutes(ctx, &wg, apiClient, apiRoutesChan)

				wg.Add(1)
				go fetchControllerRoutes(ctx, &wg, goBGPClient, controllerRoutesChan)

				wg.Add(1)
				go synchronizeRoutes(ctx, &wg, apiRoutesChan, controllerRoutesChan, goBGPClient)

				// Graceful shutdown: Ensure events channel is closed when the context is done
				go func() {
					<-ctx.Done()  // Wait for context cancellation or deadline
					close(events) // Close the channel to signal worker goroutines to stop
				}()

				// Wait for all goroutines to finish
				fmt.Println("Updater is running. Performing tasks...")
				wg.Wait()

				// Periodically check connectivity while the program is running
				go monitorConnectivity(ctx, goBGPClient, apiClient, cancel)
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
			_, err := goBGPClient.GetBGP()
			if err != nil {
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
