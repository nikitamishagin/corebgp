package updater

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	v1 "github.com/nikitamishagin/corebgp/pkg/client/v1"
	"net"
	"sync"
)

// fetchAPIRoutes fetches all route data from the API and sends the resulting route map to the specified channel.
func fetchAPIRoutes(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, apiRoutesChan chan<- map[string]Route) {
	defer wg.Done()
	defer close(apiRoutesChan)

	fmt.Println("Fetching all routes from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		apiRoutesChan <- map[string]Route{}
		fmt.Printf("failed to fetch routes from API: %v", err)
		return
	}
	if len(announcements.Data) == 0 {
		apiRoutesChan <- map[string]Route{}
		fmt.Println("No routes found in API.")
		return
	}

	// Convert announcement information to routes
	routeMap := make(map[string]Route)
	for i := range announcements.Data {
		// Parse announced IP
		ip, ipNet, err := net.ParseCIDR(announcements.Data[i].Addresses.AnnouncedIP)
		if err != nil {
			fmt.Printf("error parsing announced IP: %v\n", err)
			continue
		}
		mask, _ := ipNet.Mask.Size()

		// Convert announcement information to routes
		for j := range announcements.Data[i].Status {
			// Get healthy next hops from announcement status
			if announcements.Data[i].Status[j].Health {
				// Make route object
				route := Route{
					Prefix:       ip.String(),
					PrefixLength: uint32(mask),
					NextHop:      announcements.Data[i].Status[j].NextHop,
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
	return
}

// fetchControllerRoutes fetches all controller routes from the GoBGP server and sends them to the provided channel.
func fetchControllerRoutes(ctx context.Context, wg *sync.WaitGroup, goBGPClient *GoBGPClient, controllerRoutesChan chan<- map[string]Route) {
	defer wg.Done()
	defer close(controllerRoutesChan)

	fmt.Println("Fetching all routes from GoBGP...")

	// Fetch all routes from the GoBGP server using a wildcard prefix ("0.0.0.0/0").
	routes, err := goBGPClient.ListPath(ctx, []string{"0.0.0.0/0"})
	if err != nil {
		// If there's an error, send an empty route map and return the error.
		controllerRoutesChan <- map[string]Route{}
		fmt.Printf("failed to fetch routes from GoBGP: %v", err)
		return
	}

	// If no routes are returned, log the information and send an empty map to the channel.
	if len(routes) == 0 {
		controllerRoutesChan <- map[string]Route{}
		fmt.Println("No routes found in GoBGP.")
		return
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
	return
}

// synchronizeRoutes synchronizes BGP routes between an API channel and a controller channel using a GoBGP client.
func synchronizeRoutes(ctx context.Context, wg *sync.WaitGroup, apiRoutesChan <-chan map[string]Route, controllerRoutesChan <-chan map[string]Route, goBGPClient *GoBGPClient) {
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
			err := goBGPClient.DeletePath(ctx, toRemove[i])
			if err != nil {
				fmt.Printf("failed to remove route: %v\n", err)
			}
		}
	}

	// Process addition of routes to GoBGP.
	if len(toAdd) > 0 {
		fmt.Printf("Adding %d routes to GoBGP...\n", len(toAdd))

		err := goBGPClient.AddPaths(ctx, toAdd)
		if err != nil {
			fmt.Printf("failed to add routes: %v\n", err)
		}
	}

	fmt.Println("Route synchronization completed.")
	return
}

func watchAnnouncements(ctx context.Context, cancel context.CancelFunc, apiClient *v1.APIClient, routeUpdateChan chan<- RouteUpdate) {
	defer close(routeUpdateChan)

	err := apiClient.WatchAnnouncements(ctx, func(event model.WatchEvent) {
		select {
		case <-ctx.Done():
			fmt.Println("Context canceled, stopping watchAnnouncements...")
			return
		default:
			for i := range event.Data.NextHops {
				if !event.Data.Status[i].Health {
					continue
				}

				ip, ipNet, err := net.ParseCIDR(event.Data.Addresses.AnnouncedIP)
				if err != nil {
					fmt.Printf("error parsing announced IP: %v\n", err)
					continue
				}
				mask, _ := ipNet.Mask.Size()

				routeUpdate := RouteUpdate{
					Type: event.Type,
					Route: Route{
						Prefix:       ip.String(),
						PrefixLength: uint32(mask),
						NextHop:      event.Data.NextHops[i],
						Origin:       0,
						Identifier:   uint32(i),
					},
				}
				routeUpdateChan <- routeUpdate
			}
		}
	})

	if err != nil {
		fmt.Printf("error while watching announcements: %v\n", err)
		cancel() // Cancel the context in case of an error
	}
}

func routesHanding(ctx context.Context, wg *sync.WaitGroup, client *GoBGPClient, routeUpdates <-chan RouteUpdate) {
	defer wg.Done()

	for routeUpdate := range routeUpdates {
		fmt.Printf("Processing event: type=%s, prefix=%s, next-hop=%s\n", routeUpdate.Type, routeUpdate.Route.Prefix, routeUpdate.Route.NextHop)

		// Handle the event based on the Type
		switch routeUpdate.Type {
		case model.EventAdded:
			err := client.AddPaths(ctx, []Route{routeUpdate.Route})
			if err != nil {
				fmt.Printf("failed to add route %s via %v: %v\n", routeUpdate.Route.Prefix, routeUpdate.Route.NextHop, err)
			}
		case model.EventUpdated:
			err := client.UpdatePaths(ctx, []Route{routeUpdate.Route})
			if err != nil {
				fmt.Printf("failed to update route %s via %v: %v\n", routeUpdate.Route.Prefix, routeUpdate.Route.NextHop, err)
			}
		case model.EventDeleted:
			err := client.DeletePath(ctx, routeUpdate.Route)
			if err != nil {
				fmt.Printf("failed to delete route %s via %v: %v\n", routeUpdate.Route.Prefix, routeUpdate.Route.NextHop, err)
			}
		default:
			fmt.Printf("unrecognized event type: %s\n", routeUpdate.Type)
		}
	}
}
