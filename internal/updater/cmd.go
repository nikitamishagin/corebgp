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

			// Initialize the new GoBGP client
			goBGPClient, err := NewGoBGPClient(&config.GoBGPEndpoint, &config.GoBGPCACert, &config.GoBGPClientCert, &config.GoBGPClientKey)
			if err != nil {
				return err
			}
			defer goBGPClient.Close()

			// TODO: Implement GoBGP configuration checking
			_, err = goBGPClient.GetBGP()
			if err != nil {
				return err
			}

			// TODO: Implement reconnection

			// Initialize the CoreBGP API client
			apiClient := v1.NewAPIClient(&config.APIEndpoint, time.Second*5)

			// Check if CoreBGP API server is healthy
			err = apiClient.HealthCheck(ctx)
			if err != nil {
				return err
			}

			// Create channels to synchronize data
			announcementsChan := make(chan []model.Announcement)
			routesChan := make(chan []model.Route)

			// Create a WaitGroup to manage goroutines
			var wg sync.WaitGroup

			// TODO: Replace events channel to routes channel
			// Create a channel to process events
			events := make(chan model.Event, 100) // Buffered channel to handle bursts of events
			defer close(events)

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

			// Goroutine for fetching all announcements
			wg.Add(1)
			go func(ctx context.Context) {
				defer wg.Done()

				fmt.Println("Fetching all announcements from CoreBGP API...")
				announcements, err := apiClient.GetAllAnnouncements(ctx)
				if err != nil {
					fmt.Printf("Failed to fetch announcements: %v\n", err)
					cancel()
					return
				}

				announcementsChan <- announcements
				close(announcementsChan)
			}(ctx)

			// Goroutine for fetching all routes from GoBGP
			wg.Add(1)
			go func(ctx context.Context) {
				defer wg.Done()
				fmt.Println("Fetching all routes from GoBGP...")
				routes, err := goBGPClient.ListPaths(ctx)
				if err != nil {
					fmt.Printf("Failed to fetch routes: %v\n", err)
					cancel()
					return
				}
				routesChan <- routes
				close(routesChan)
			}(ctx)

			// TODO: Add comparing existed routes with API routes

			// TODO: Add update GoBGP routes

			// Graceful shutdown: Ensure events channel is closed when the context is done
			go func() {
				<-ctx.Done()  // Wait for context cancellation or deadline
				close(events) // Close the channel to signal worker goroutines to stop
			}()
			fmt.Println("Updater is running. Listening for events and performing tasks...")

			// Wait for all goroutines to finish
			wg.Wait()

			return nil
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
