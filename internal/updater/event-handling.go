package updater

import (
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
)

func handleAnnouncementEvent(client *GoBGPClient, event *model.Event) error {
	// Log the event being processed
	fmt.Printf("Processing event: type=%s, address=%s, next-hops=%v\n", event.Type, event.Announcement.Addresses.AnnouncedIP, event.Announcement.NextHops)

	// Handle the event based on the Type
	switch event.Type {
	case model.EventAdded:
		// Add route (only one next hop for test)
		err := client.AddPath(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops[0].IP)
		if err != nil {
			return fmt.Errorf("failed to add route %s via %v: %w", event.Announcement.Addresses.AnnouncedIP, event.Announcement.NextHops, err)
		}
	//case model.EventUpdated:
	//	// Update announcement (update route)
	//	err := client.UpdatePath(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops[0].IP)
	//	if err != nil {
	//		return fmt.Errorf("failed to update route %s/%d: %w",
	//			event.Announcement.Addresses.AnnouncedIP, 32, err)
	//	}
	case model.EventDeleted:
		// Delete announcement (remove route)
		err := client.DeletePath(event.Announcement.Addresses.AnnouncedIP, 32, event.Announcement.NextHops[0].IP)
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
