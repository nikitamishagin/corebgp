package checker

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"sync"
)

func watchAnnouncements(ctx context.Context, cancel context.CancelFunc, apiClient *v1.APIClient, healthCheckChan chan<- HealthCheck) {
	defer close(healthCheckChan)

	err := apiClient.WatchAnnouncements(ctx, func(event model.WatchEvent) {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled, stopping watchAnnouncements...")
			return
		default:
			for i := range event.Data.NextHops {
				HealthCheck := HealthCheck{
					NextHop:       event.Data.NextHops[i],
					Path:          event.Data.HealthCheck.Path,
					Port:          event.Data.HealthCheck.Port,
					Method:        event.Data.HealthCheck.Method,
					CheckInterval: event.Data.HealthCheck.CheckInterval,
					Timeout:       event.Data.HealthCheck.Timeout,
					Delay:         0,
				}
				healthCheckChan <- HealthCheck
			}
		}
	})

	if err != nil {
		fmt.Printf("error while watching announcements: %v\n", err)
		cancel() // Cancel the context in case of an error
	}
}

func fetchHealthChecks(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, healthCheckMapChan chan<- map[string]HealthCheck) {
	defer wg.Done()
	defer close(healthCheckMapChan)

	fmt.Println("Fetching all health checks from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		healthCheckMapChan <- map[string]HealthCheck{}
		fmt.Printf("failed to fetch health checks from API: %v", err)
		return
	}
	if len(announcements.Data) == 0 {
		healthCheckMapChan <- map[string]HealthCheck{}
		fmt.Println("No health checks found in API.")
		return
	}

	// Convert announcement information to health checks
	healthCheckMap := make(map[string]HealthCheck)
	for i := range announcements.Data {
		// Convert announcement information to health check object
		for j := range announcements.Data[i].NextHops {
			// Make health check object
			healthCheck := HealthCheck{
				NextHop:       announcements.Data[i].NextHops[j],
				Path:          announcements.Data[i].HealthCheck.Path,
				Port:          announcements.Data[i].HealthCheck.Port,
				Method:        announcements.Data[i].HealthCheck.Method,
				CheckInterval: announcements.Data[i].HealthCheck.CheckInterval,
				Timeout:       announcements.Data[i].HealthCheck.Timeout,
				Delay:         0,
			}

			// Create a new key and write to map
			key := fmt.Sprintf("%s:%d%s_%s", healthCheck.NextHop, healthCheck.Port, healthCheck.Path, healthCheck.Method)
			healthCheckMap[key] = healthCheck
		}
	}

	// Send the constructed health check map to the provided channel.
	healthCheckMapChan <- healthCheckMap
	return
}

// TODO: Implement calculate function
func calculateDelay(ctx context.Context, cancel context.CancelFunc, healthCheck HealthCheck) {

}
