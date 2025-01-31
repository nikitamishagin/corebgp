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

func fetchAPIRoutes(ctx context.Context, wg *sync.WaitGroup, apiClient *v1.APIClient, apiRoutesChan chan<- []model.Route) {
	defer wg.Done()
	defer close(apiRoutesChan)

	fmt.Println("Fetching all routes from API...")

	// Get all announcements from CoreBGP API
	announcements, err := apiClient.GetAllAnnouncements(ctx)
	if err != nil {
		fmt.Printf("Error fetching routes from API: %v\n", err)
		apiRoutesChan <- []model.Route{}
		return
	}
	if len(announcements) == 0 {
		fmt.Println("No routes found in API.")
		apiRoutesChan <- []model.Route{}
		return
	}

	// Define max capacity for a slice of routes
	var count = 0
	for i := range announcements {
		count += len(announcements[i].Status.Details)
	}

	// Convert announcement information to routes
	allRoutes := make([]model.Route, 0, count)
	for i := range announcements {
		routes, err := announcementToRoutes(&announcements[i])
		if err != nil {
			fmt.Printf("Error converting announcement to routes: %v\n", err)
			continue
		}
		allRoutes = append(allRoutes, routes...)
	}
	apiRoutesChan <- allRoutes
}

// TODO: Refactor fetchControllerRoutes func
func fetchControllerRoutes(ctx context.Context, wg *sync.WaitGroup, goBGPClient *GoBGPClient, controllerRoutesChan chan<- []model.Route) {
	defer wg.Done()
	defer close(controllerRoutesChan)

	fmt.Println("Fetching all routes from GoBGP...")
	var allRoutes []Route

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

	var apiRoutes []Route
	var controllerRoutes []Route

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

func announcementToRoutes(announcement *model.Announcement) ([]model.Route, error) {
	// Parse announced IP
	ip, ipNet, err := net.ParseCIDR(announcement.Addresses.AnnouncedIP)
	if err != nil {
		return nil, err
	}
	mask, _ := ipNet.Mask.Size()

	// Convert announcement information to routes
	routes := make([]model.Route, 0, len(announcement.Status.Details))
	for i := range announcement.Status.Details {
		var nextHop string

		// Get healthy next hops from announcement status
		if announcement.Status.Details[i].Status == "health" {
			nextHop = announcement.Status.Details[i].Host
		}

		// Make route
		route := model.Route{
			Prefix:       ip.String(),
			PrefixLength: uint32(mask),
			NextHop:      nextHop,
			Origin:       0,
			Identifier:   i,
		}
		routes = append(routes, route)
	}

	return routes, nil
}
