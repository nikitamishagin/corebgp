package model

// Route represents a GoBGP route with its prefix, prefix length, next hop IPs, origin, and a unique identifier.
type Route struct {
	Prefix       string   // The IP prefix for the route
	PrefixLength uint32   // The prefix length (subnet mask)
	NextHops     []string // List of next-hop IP addresses
	Origin       uint32   // Origin attribute, e.g., 0 for IGP
	Identifier   uint32   // Unique identifier for the path
}
