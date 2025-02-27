package checker

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"github.com/spf13/cobra"
	"sync"
	"time"
)

func RootCmd() *cobra.Command {
	var config model.CheckerConfig
	var cmd = &cobra.Command{
		Use:   "checker",
		Short: "CoreBGP checker",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement checker logic
			ctx, cancel := context.WithCancel(cmd.Context())

			// Initialize the CoreBGP API client
			apiClient := v1.NewAPIClient(&config.APIEndpoint, time.Second*5)
			if err := apiClient.HealthCheck(ctx); err != nil {
				fmt.Printf("Failed to connect to CoreBGP API: %v. Retrying...\n", err)
			}

			taskUpdatesChan := make(chan TaskUpdate, 100)

			go watchAnnouncements(ctx, cancel, apiClient, taskUpdatesChan)

			// Create channels for routes
			tasksMapChan := make(chan map[string]Task, 1)

			// Create a WaitGroup to manage goroutines
			var wg sync.WaitGroup

			wg.Add(1)
			go fetchTasks(ctx, &wg, apiClient, tasksMapChan)

			return nil
		},
	}

	cmd.Flags().StringVar(&config.APIEndpoint, "api-endpoint", "http://localhost:8080", "URL of the API server")
	cmd.Flags().StringVar(&config.Zone, "zone", "default", "Zone name")
	cmd.Flags().StringVar(&config.LivenessTimeout, "liveness-timeout", "10s", "Liveness timeout")
	cmd.Flags().StringVar(&config.LogPath, "log-path", "/var/log/corebgp/checker.log", "Path to the log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}
