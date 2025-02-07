package updater

import (
	"context"
	"fmt"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"net"
	"sync"
)

// fetchAPIRoutes fetches all route data from the API and sends the resulting route map to the specified channel.
func fetchAPIRoutes(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, apiRoutesChan chan<- map[string]Route) error {
	defer wg.Done()
	defer close(apiRoutesChan)

	fmt.Println("Fetching all routes from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		apiRoutesChan <- map[string]Route{}
		return fmt.Errorf("failed to fetch routes from API: %w", err)
	}
	if len(announcements) == 0 {
		apiRoutesChan <- map[string]Route{}
		fmt.Println("No routes found in API.")
		return nil
	}

	// Convert announcement information to routes
	routeMap := make(map[string]Route)
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
				route := Route{
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

	// Send the constructed route map to the provided channel.
	apiRoutesChan <- routeMap
	return nil
}

// fetchControllerRoutes fetches all controller routes from the GoBGP server and sends them to the provided channel.
func fetchControllerRoutes(ctx context.Context, wg *sync.WaitGroup, goBGPClient *GoBGPClient, controllerRoutesChan chan<- map[string]Route) error {
	defer wg.Done()
	defer close(controllerRoutesChan)

	fmt.Println("Fetching all routes from GoBGP...")

	// Fetch all routes from the GoBGP server using a wildcard prefix ("0.0.0.0/0").
	routes, err := goBGPClient.ListPath(ctx, []string{"0.0.0.0/0"})
	if err != nil {
		// If there's an error, send an empty route map and return the error.
		controllerRoutesChan <- map[string]Route{}
		return fmt.Errorf("failed to fetch routes from GoBGP: %w", err)
	}

	// If no routes are returned, log the information and send an empty map to the channel.
	if len(routes) == 0 {
		controllerRoutesChan <- map[string]Route{}
		fmt.Println("No routes found in GoBGP.")
		return nil
	}

	// Initialize a map to store routes with unique keys.
	routeMap := make(map[string]Route)
	for i := range routes {
		// Create a unique key for each route using its prefix, prefix length, and next hop.
		key := fmt.Sprintf("%s/%d-%v", routes[i].Prefix, routes[i].PrefixLength, routes[i].NextHop)

		routeMap[key] = routes[i]
	}

	// Send the constructed route map to the provided channel.
	controllerRoutesChan <- routeMap
	return nil
}

// synchronizeRoutes synchronizes BGP routes between an API channel and a controller channel using a GoBGP client.
func synchronizeRoutes(ctx context.Context, wg *sync.WaitGroup, apiRoutesChan <-chan map[string]Route, controllerRoutesChan <-chan map[string]Route, goBGPClient *GoBGPClient) error {
	defer wg.Done()

	var (
		apiRouteMap, controllerRouteMap map[string]Route
		apiOk, controllerOk             bool
	)

	// Loop until both apiRoutesChan and controllerRoutesChan are closed and all data is received.
	for !apiOk || !controllerOk {
		select {
		case apiRoutes, open := <-apiRoutesChan:
			if open {
				apiRouteMap = apiRoutes
			} else {
				apiOk = true
			}

		// Receive data from the controller routes channel.
		case controllerRoutes, open := <-controllerRoutesChan:
			if open {
				controllerRouteMap = controllerRoutes
			} else {
				controllerOk = true
			}
		}
	}

	// Variables to store routes that need to be added to or removed from GoBGP.
	var toAdd, toRemove []Route

	// Compare the API route map with the controller route map.
	// Identify routes to add (present in the API but not in the controller).
	for key, route := range apiRouteMap {
		if _, exists := controllerRouteMap[key]; !exists {
			toAdd = append(toAdd, route)
		} else {
			// Remove matching routes from the controller map for further processing.
			delete(controllerRouteMap, key)
		}
	}

	// Identify routes to remove (those still left in the controller map).
	for i := range controllerRouteMap {
		toRemove = append(toRemove, controllerRouteMap[i])
	}

	// Process removal of routes from GoBGP.
	if len(toRemove) > 0 {
		fmt.Printf("Removing %d routes from GoBGP...\n", len(toRemove))

		for i := range toRemove {
			// Attempt to remove each route from GoBGP.
			err := goBGPClient.DeletePath(ctx, toRemove[i].Prefix, []string{toRemove[i].NextHop}, toRemove[i].PrefixLength, toRemove[i].Origin, toRemove[i].Identifier)
			if err != nil {
				fmt.Printf("Failed to remove routes: %v\n", err)
			}
		}
	}

	// Process addition of routes to GoBGP.
	if len(toAdd) > 0 {
		fmt.Printf("Adding %d routes to GoBGP...\n", len(toAdd))

		for i := range toAdd {
			// Attempt to add each route to GoBGP.
			err := goBGPClient.AddPaths(ctx, toAdd[i].Prefix, []string{toAdd[i].NextHop}, toAdd[i].PrefixLength, toAdd[i].Origin, toAdd[i].Identifier)
			if err != nil {
				fmt.Printf("Failed to add routes: %v\n", err)
			}
		}
	}

	fmt.Println("Route synchronization completed.")
	return nil
}
