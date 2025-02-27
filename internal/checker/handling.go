package checker

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"sync"
)

func watchAnnouncements(ctx context.Context, cancel context.CancelFunc, apiClient *v1.APIClient, taskUpdatesChan chan<- TaskUpdate) {
	defer close(taskUpdatesChan)

	err := apiClient.WatchAnnouncements(ctx, func(event model.WatchEvent) {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled, stopping watchAnnouncements...")
			return
		default:
			for i := range event.Data.NextHops {
				taskUpdate := TaskUpdate{
					Type: event.Type,
					Tasks: Task{
						NextHop:       event.Data.NextHops[i],
						Path:          event.Data.HealthCheck.Path,
						Port:          event.Data.HealthCheck.Port,
						Method:        event.Data.HealthCheck.Method,
						CheckInterval: event.Data.HealthCheck.CheckInterval,
						Timeout:       event.Data.HealthCheck.Timeout,
						Delay:         0,
					},
				}
				taskUpdatesChan <- taskUpdate
			}
		}
	})

	if err != nil {
		fmt.Printf("error while watching announcements: %v\n", err)
		cancel() // Cancel the context in case of an error
	}
}

func fetchTasks(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, tasksMapChan chan<- map[string]Task) {
	defer wg.Done()
	defer close(tasksMapChan)

	fmt.Println("Fetching all health checks from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		tasksMapChan <- map[string]Task{}
		fmt.Printf("failed to fetch health checks from API: %v", err)
		return
	}
	if len(announcements.Data) == 0 {
		tasksMapChan <- map[string]Task{}
		fmt.Println("No health checks found in API.")
		return
	}

	// Convert announcement information to health checks
	tasksMap := make(map[string]Task)
	for i := range announcements.Data {
		// Convert announcement information to health check object
		for j := range announcements.Data[i].NextHops {
			// Make health check object
			task := Task{
				NextHop:       announcements.Data[i].NextHops[j],
				Path:          announcements.Data[i].HealthCheck.Path,
				Port:          announcements.Data[i].HealthCheck.Port,
				Method:        announcements.Data[i].HealthCheck.Method,
				CheckInterval: announcements.Data[i].HealthCheck.CheckInterval,
				Timeout:       announcements.Data[i].HealthCheck.Timeout,
				Delay:         0,
			}

			// Create a new key and write to map
			key := fmt.Sprintf("%s:%d%s_%s", task.NextHop, task.Port, task.Path, task.Method)
			tasksMap[key] = task
		}
	}

	// Send the constructed health check map to the provided channel.
	tasksMapChan <- tasksMap
	return
}

// TODO: Implement calculate function
func calculateDelay(ctx context.Context, cancel context.CancelFunc, healthCheck Task) {

}

func runningTasks(ctx context.Context, wg *sync.WaitGroup, tasksMapChan <-chan map[string]Task, activeTasksChan chan<- map[string]context.CancelFunc) {
	defer wg.Done()

	tasksMap := <-tasksMapChan
	for key, task := range tasksMap {
		ctx, cancel := context.WithCancel(ctx)

		go runTask(ctx, cancel, task)
		activeTasksChan <- map[string]context.CancelFunc{key: cancel}
	}
}

// TODO: Implement running task.
func runTask(ctx context.Context, cancel context.CancelFunc, task Task) {
	defer cancel()
}
