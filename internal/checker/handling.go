package checker

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
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

// TODO: Implement calculate function
func calculateDelay(ctx context.Context, cancel context.CancelFunc, healthCheck HealthCheck) {

}
