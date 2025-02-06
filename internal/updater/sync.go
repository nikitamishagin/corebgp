package updater

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
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

func fetchControllerRoutes(ctx context.Context, wg *sync.WaitGroup, goBGPClient *GoBGPClient, controllerRoutesChan chan<- map[string]model.Route) error {
	defer wg.Done()
	defer close(controllerRoutesChan)

	fmt.Println("Fetching all routes from GoBGP...")

	routes, err := goBGPClient.ListPath(ctx, []string{"0.0.0.0/0"})
	if err != nil {
		controllerRoutesChan <- map[string]model.Route{}
		return fmt.Errorf("failed to fetch routes from GoBGP: %w", err)
	}
	if len(routes) == 0 {
		controllerRoutesChan <- map[string]model.Route{}
		fmt.Println("No routes found in GoBGP.")
		return nil
	}

	routeMap := make(map[string]model.Route)
	for i := range routes {
		// Create a new key and write to map
		key := fmt.Sprintf("%s/%d-%v", routes[i].Prefix, routes[i].PrefixLength, routes[i].NextHop)
		routeMap[key] = routes[i]
	}
	return nil
}

// TODO: Complete synchronizeRoutes function
func synchronizeRoutes(ctx context.Context, wg *sync.WaitGroup, apiRoutesChan <-chan map[string]model.Route, controllerRoutesChan <-chan map[string]model.Route, goBGPClient *GoBGPClient) error {
	defer wg.Done()

	select {
	case apiRouteMap, ok := <-apiRoutesChan:
		if !ok {
			return fmt.Errorf("apiRoutesChan closed unexpectedly")
		}

	case controllerRouteMap, ok := <-controllerRoutesChan:
		if !ok {
			return fmt.Errorf("controllerRoutesChan closed unexpectedly")
		}
	}

	apiRouteMap := <-apiRoutesChan
	controllerRouteMap := <-controllerRoutesChan

	var toAdd, toRemove []model.Route

	for key, route := range apiRouteMap {
		if _, exists := controllerRouteMap[key]; !exists {
			goBGPClient.AddPaths(ctx)
		} else {
			delete(controllerRouteMap, key)
		}
	}

	for _, route := range controllerRouteMap {
		toRemove = append(toRemove, route)
	}

	if len(toRemove) > 0 {
		fmt.Printf("Removing %d routes from GoBGP...\n", len(toRemove))
		err := goBGPClient.RemoveRoutes(ctx, toRemove)
		if err != nil {
			fmt.Printf("Failed to remove routes: %v\n", err)
		}
	}

	if len(toAdd) > 0 {
		fmt.Printf("Adding %d routes to GoBGP...\n", len(toAdd))
		err := goBGPClient.AddRoutes(ctx, toAdd)
		if err != nil {
			fmt.Printf("Failed to add routes: %v\n", err)
		}
	}

	fmt.Println("Route synchronization completed.")
	return nil
}
