package updater

// Route represents a GoBGP route with its prefix, prefix length, next hop IP, origin, and a unique identifier.
type Route struct {
	Prefix       string // The IP prefix for the route
	PrefixLength uint32 // The prefix length (subnet mask)
	NextHop      string // The next-hop IP address
	Origin       uint32 // Origin attribute, e.g., 0 for IGP
	Identifier   uint32 // Unique identifier for the path
}
