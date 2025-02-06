package model

// EventType defines the type of event such as added, updated or deleted.
type EventType string

const (
	EventAdded   EventType = "added"   // EventAdded represents the event type for adding a new announcement.
	EventUpdated EventType = "updated" // EventUpdated represents the event type for updating an existing announcement.
	EventDeleted EventType = "deleted" // EventDeleted represents the event type for deleting an existing announcement.
)

// Event represents a BGP announcement event, encapsulating the type of action and the specific announcement.
type Event struct {
	Type         EventType    `json:"type"`         // Action specifies the type of event: add, update, or delete.
	Announcement Announcement `json:"announcement"` // Announcement is the BGP announcement data associated with the event.
}

// APIResponse represents a standard response structure for API calls.
type APIResponse struct {
	Status  string      `json:"status"`  // Status indicates the operation outcome: success or error.
	Message string      `json:"message"` // Message provides additional details about the result of the API call.
	Data    interface{} `json:"data"`    // Data contains the response payload, which can vary depending on the endpoint.
}

// Announcement represents a BGP routing configuration, including metadata, addresses, next-hop details, health checks, and status.
type Announcement struct {
	Meta        Meta        `json:"meta"`         // Meta represents metadata information including a descriptive name and associated project for a BGP announcement.
	Addresses   Addresses   `json:"addresses"`    // Addresses represents a collection of network-related data, including subnets, zone, and announcing ip.
	NextHops    []string    `json:"next-hops"`    // NextHops represents a collection of next-hop IP addresses used for routing purposes.
	HealthCheck HealthCheck `json:"health-check"` // HealthCheck represents the configuration and parameters for performing health checks on next hops.
	Status      []Status    `json:"statuses"`     // Status represents the state of a health check including results, details, and timing information.
}

// Meta represents metadata information including a descriptive name and associated project for a BGP announcement.
type Meta struct {
	Name    string `json:"name"`    // Name specifies the descriptive name for the BGP announce.
	Project string `json:"project"` // Project specifies the project associated with the BGP announce.
}

// Addresses represents a collection of network-related data, including subnets, zone, and announcing ip.
type Addresses struct {
	SourceSubnets Subnet `json:"source-subnets"` // SourceSubnets specifies the subnet from which the announced address should be obtained (IPAM).
	Zone          string `json:"zone"`           // Zone specifies the geographical or logical zone associated with the addresses.
	AnnouncedIP   string `json:"announced-ip"`   // AnnouncedIP specifies the IP address being announced for routing purposes.
}

// Subnet represents a network subnet with an IP address and subnet mask.
type Subnet struct {
	IP   string `json:"ip"`   // IP represents the IP address in string format.
	Mask uint8  `json:"mask"` // Mask represents the subnet mask as an unsigned 8-bit integer.
}

// HealthCheck is a configuration for performing health checks on the next hop.
type HealthCheck struct {
	Path          string `json:"path"`     // Path specifies the endpoint to be used for the health check process.
	Port          int    `json:"port"`     // Port specifies the port number to be used for the health check process.
	Method        string `json:"method"`   // Method specifies the HTTP method to be used for the health check process.
	CheckInterval int    `json:"interval"` // CheckInterval specifies the interval in seconds between consecutive health check attempts.
	Timeout       int    `json:"timeout"`  // Timeout specifies the duration in milliseconds before a health check request times out.
}

// Status represents the state of a health check including results, details, and timing information.
type Status struct {
	NextHop   string `json:"next-hop"`  // NextHop specifies the IP address to which traffic should be routed as the next hop in the network path.
	Health    bool   `json:"health"`    // Health indicates whether the health check passed (true) or failed (false).
	Code      int    `json:"code"`      // Code is the health check HTTP response status codes.
	Message   string `json:"msg"`       // Message provides additional details or context about the health check result.
	Timestamp string `json:"timestamp"` // Timestamp represents the time at which the status was recorded in ISO 8601 format.
}
