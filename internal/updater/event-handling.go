package updater

import "fmt"

func handleAnnouncementEvent(client *GoBGPClient, event map[string]interface{}) error {
	// Ensure the event has a recognized format and type
	action, ok := event["action"].(string)
	if !ok {
		return fmt.Errorf("event is missing required 'action' field")
	}

	// Parse the prefix and prefix length
	prefix, ok := event["prefix"].(string)
	if !ok {
		return fmt.Errorf("event is missing 'prefix' field")
	}
	prefixLengthFloat, ok := event["prefix_length"].(float64) // JSON unmarshals numbers to float64
	if !ok {
		return fmt.Errorf("event is missing 'prefix_length' field")
	}
	prefixLength := uint32(prefixLengthFloat)
	nextHop, ok := event["nexthop"].(string)
	if !ok {
		return fmt.Errorf("event is missing 'nexthop' field")
	}

	// Log the event being processed
	fmt.Printf("Processing event: action=%s, prefix=%s/%d\n", action, prefix, prefixLength)

	// Handle actions: "add", "update", and "delete"
	switch action {
	case "add":
		err := client.AddPath(prefix, prefixLength, nextHop)
		if err != nil {
			return fmt.Errorf("failed to add route %s/%d: %w", prefix, prefixLength, err)
		}
	// TODO: Implement update route
	//case "update":
	//	err := client.UpdatePath(prefix, prefixLength)
	//	if err != nil {
	//		return fmt.Errorf("failed to update route %s/%d: %w", prefix, prefixLength, err)
	//	}

	case "delete":
		err := client.DeletePath(prefix, prefixLength, nextHop)
		if err != nil {
			return fmt.Errorf("failed to delete route %s/%d: %w", prefix, prefixLength, err)
		}

	default:
		return fmt.Errorf("unrecognized action: %s", action)
	}

	return nil
}
