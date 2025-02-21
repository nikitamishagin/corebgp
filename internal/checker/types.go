package checker

// HealthCheck represents the configuration for performing health checks on a service or endpoint.
type HealthCheck struct {
	NextHop       string `json:"nextHop"`  // NextHop specifies the address of the next hop for routing or forwarding traffic.
	Path          string `json:"path"`     // Path specifies the endpoint to be used for the health check process.
	Port          int    `json:"port"`     // Port specifies the port number to be used for the health check process.
	Method        string `json:"method"`   // Method specifies the HTTP method to be used for the health check process.
	CheckInterval int    `json:"interval"` // CheckInterval specifies the interval in seconds between consecutive health check attempts.
	Timeout       int    `json:"timeout"`  // Timeout specifies the duration in milliseconds before a health check request times out.
	Delay         int    `json:"delay"`    // Delay specifies the time in seconds to wait before initiating the first health check.
}
