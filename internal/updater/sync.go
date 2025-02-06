package updater

import (
	"context"
	"errors"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"io"
	"net"
	"sync"
)

func fetchAPIRoutes(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, apiRoutesChan chan<- map[string]model.Route) error {
	defer wg.Done()
	defer close(apiRoutesChan)

	fmt.Println("Fetching all routes from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		apiRoutesChan <- map[string]model.Route{}
		return fmt.Errorf("failed to fetch routes from API: %w", err)
	}
	if len(announcements) == 0 {
		apiRoutesChan <- map[string]model.Route{}
		fmt.Println("No routes found in API.")
		return nil
	}

	// Convert announcement information to routes
	routeMap := make(map[string]model.Route)
	for i := range announcements {
		// Parse announced IP
		ip, ipNet, err := net.ParseCIDR(announcements[i].Addresses.AnnouncedIP)
		if err != nil {
			fmt.Printf("error parsing announced IP: %v\n", err)
			continue
		}
		mask, _ := ipNet.Mask.Size()

		// Convert announcement information to routes
		for j := range announcements[i].Status {
			// Get healthy next hops from announcement status
			if announcements[i].Status[j].Health {
				// Make route object
				route := model.Route{
					Prefix:       ip.String(),
					PrefixLength: uint32(mask),
					NextHop:      announcements[i].Status[j].NextHop,
					Origin:       0,
					Identifier:   uint32(j),
				}

				// Create a new key and write to map
				key := fmt.Sprintf("%s/%d-%v", route.Prefix, route.PrefixLength, route.NextHop)
				routeMap[key] = route
			}
		}
	}
	apiRoutesChan <- routeMap
	return nil
}

// TODO: Refactor fetchControllerRoutes func
func fetchControllerRoutes(ctx context.Context, wg *sync.WaitGroup, goBGPClient *GoBGPClient, controllerRoutesChan chan<- []model.Route) {
	defer wg.Done()
	defer close(controllerRoutesChan)

	fmt.Println("Fetching all routes from GoBGP...")
	var allRoutes []model.Route

	for {
		paths, err := goBGPClient.ListPath(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // No more data
			}
			fmt.Printf("Error fetching routes from GoBGP: %v\n", err)
			return
		}
		allRoutes = append(allRoutes, batch...)
	}

	controllerRoutesChan <- allRoutes
}

// TODO: Refactor synchronizeRoutes func
func synchronizeRoutes(ctx context.Context, wg *sync.WaitGroup, apiRoutesChan <-chan []model.Route, controllerRoutesChan <-chan []model.Route, goBGPClient *GoBGPClient) {
	defer wg.Done()

	var apiRoutes []model.Route
	var controllerRoutes []model.Route

	// Receive routes from both channels
	select {
	case apiRoutes = <-apiRoutesChan:
	case <-ctx.Done():
		fmt.Println("Context canceled before receiving API routes.")
		return
	}

	select {
	case controllerRoutes = <-controllerRoutesChan:
	case <-ctx.Done():
		fmt.Println("Context canceled before receiving GoBGP routes.")
		return
	}

	// Perform synchronization
	fmt.Println("Starting route synchronization...")
	if err := syncRoutes(apiRoutes, controllerRoutes, goBGPClient); err != nil {
		fmt.Printf("Failed to synchronize routes: %v\n", err)
	}
}
