package updater

import (
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
)

//TODO: Replace event struct to route struct

// handleAnnouncementEvent processes a BGP announcement event to add, update, or delete routes based on the event type.
// The function interacts with a GoBGPClient instance to update BGP routes accordingly.
// If the event type is unrecognized or an error occurs during processing, it returns an error.
func handleAnnouncementEvent(client *GoBGPClient, event *model.Event) error {
	// Log the event being processed
	fmt.Printf("Processing event: type=%s, address=%s, next-hops=%v\n", event.Type, event.Announcement.Addresses.AnnouncedIP, event.Announcement.NextHops)

	// Handle the event based on the Type
	switch event.Type {
	case model.EventAdded:
		// Add routes
		err := client.AddPaths(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops)
		if err != nil {
			return fmt.Errorf("failed to add route %s via %v: %w", event.Announcement.Addresses.AnnouncedIP, event.Announcement.NextHops, err)
		}
	case model.EventUpdated:
		// Update routes
		err := client.UpdatePath(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops)
		if err != nil {
			return fmt.Errorf("failed to update route %s/%d: %w",
				event.Announcement.Addresses.AnnouncedIP, 32, err)
		}
	case model.EventDeleted:
		// Delete routes
		err := client.DeletePath(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops)
		if err != nil {
			return fmt.Errorf("failed to delete route %s/%d: %w",
				event.Announcement.Addresses.AnnouncedIP, 32, err)
		}
	default:
		// Unrecognized event type
		return fmt.Errorf("unrecognized event type: %s", event.Type)
	}

	return nil
}
