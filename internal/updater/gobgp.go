package updater

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	api "github.com/osrg/gobgp/v3/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/anypb"
	"net"
	"os"
)

// GoBGPClient is struct for manage GoBGP client
type GoBGPClient struct {
	client api.GobgpApiClient
	conn   *grpc.ClientConn
}

// NewGoBGPClient initializes the new GoBGP client
func NewGoBGPClient(endpoint, caFile, certFile, keyFile *string) (*GoBGPClient, error) {
	caCert, err := os.ReadFile(*caFile)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	cert, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load client certificate and key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	creds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}

	conn, err := grpc.Dial(*endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to GoBGP server: %w", err)
	}

	client := api.NewGobgpApiClient(conn)

	return &GoBGPClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes GoBGP API server connection
func (g *GoBGPClient) Close() {
	_ = g.conn.Close()
}

// GetBGP retrieves the current BGP configuration from the GoBGP server and returns it as a string.
func (g *GoBGPClient) GetBGP() (string, error) {
	// Create a request to retrieve the current BGP configuration
	bgpConfig, err := g.client.GetBgp(context.Background(), &api.GetBgpRequest{})
	if err != nil {
		return "", fmt.Errorf("failed to get BGP config: %w", err)
	}

	// Convert the BGP configuration to a string and return it
	return bgpConfig.String(), nil
}

// AddPaths adds multiple BGP routes (prefixes) with associated next-hops to the GoBGP server.
func (g *GoBGPClient) AddPaths(ctx context.Context, routes []Route) error {
	// Set up the stream
	stream, err := g.client.AddPathStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open AddPathStream: %w", err)
	}

	// Create a stream paths request from routes slice
	paths := make([]*api.Path, len(routes))
	for i := range routes {
		// Marshal NLRI (route information) into *anypb.Any
		nlri, err := anypb.New(&api.IPAddressPrefix{
			Prefix:    routes[i].Prefix,
			PrefixLen: routes[i].PrefixLength,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal NLRI: %w", err)
		}

		// Marshal the attributes (Pattrs) into *anypb.Any
		originAttr, err := anypb.New(&api.OriginAttribute{
			Origin: routes[i].Origin,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal origin attribute: %w", err)
		}

		// Marshal the NextHop attribute into *anypb.Any
		nextHopAttr, err := anypb.New(&api.NextHopAttribute{
			NextHop: routes[i].NextHop,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal next-hop attribute: %w", err)
		}

		// Construct each Path object
		paths[i] = &api.Path{
			Family: &api.Family{
				Afi:  api.Family_AFI_IP,
				Safi: api.Family_SAFI_UNICAST,
			},
			Nlri: nlri,
			Pattrs: []*anypb.Any{
				originAttr,
				nextHopAttr,
			},
			Identifier: routes[i].Identifier,
		}
	}

	// Construct the AddPathStreamRequest with multiple paths
	req := &api.AddPathStreamRequest{
		Paths: paths,
	}

	// Send the request through the gRPC stream
	if err := stream.Send(req); err != nil {
		return fmt.Errorf("failed to send paths in AddPathStream: %w", err)
	}

	// Close the stream and receive the server's response
	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close AddPathStream: %w", err)
	}

	return nil
}

// ListPath retrieves a list of BGP routes for the specified prefixes from the GoBGP server.
// Returns a slice of Route structures or an error.
func (g *GoBGPClient) ListPath(ctx context.Context, prefixes []string) ([]Route, error) {
	// Build the list of prefixes for the API request.
	lookupPrefixes := make([]*api.TableLookupPrefix, len(prefixes))
	for i := range prefixes {
		lookupPrefixes[i] = &api.TableLookupPrefix{
			Prefix: prefixes[i],
			Type:   api.TableLookupPrefix_LONGER, // Means searching for all more specific routes.
		}
	}

	// Call ListPath API with a prefix filter
	stream, err := g.client.ListPath(ctx, &api.ListPathRequest{
		Family: &api.Family{
			Afi:  api.Family_AFI_IP,
			Safi: api.Family_SAFI_UNICAST,
		},
		Prefixes: lookupPrefixes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list paths from GoBGP: %w", err)
	}

	// Collect routes from the stream
	var routes []Route
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error while receiving path from stream: %w", err)
		}

		// Extract the Destination from response
		dest := resp.GetDestination()
		if dest == nil {
			fmt.Printf("error while parsing destination: %v\n", resp)
			continue // Skip in case of empty destination
		}

		// Get prefix and prefix length
		prefix, ipNet, err := net.ParseCIDR(dest.Prefix)
		if err != nil {
			return nil, fmt.Errorf("error while parsing prefix: %w", err)
		}
		prefixLength, _ := ipNet.Mask.Size()

		// Loop through all paths in the Destination
		for i := range dest.Paths {
			var (
				nextHopAttr api.NextHopAttribute
				originAttr  api.OriginAttribute
			)
			// Find next hop and origin in attributes
			for _, attr := range dest.Paths[i].GetPattrs() {
				// Check the type URL of the attribute
				switch attr.GetTypeUrl() {
				case "type.googleapis.com/apipb.OriginAttribute":
					// Attempt to unmarshal the OriginAttribute
					err := attr.UnmarshalTo(&originAttr)
					if err != nil {
						fmt.Printf("error parsing origin attribute for %s prefix: %v\n", dest.Prefix, err)
					}
				case "type.googleapis.com/apipb.NextHopAttribute":
					// Attempt to unmarshal the NextHopAttribute
					err := attr.UnmarshalTo(&nextHopAttr)
					if err != nil {
						fmt.Printf("error parsing next-hop attribute for %s prefix: %v\n", dest.Prefix, err)
					}
					if err == nil && nextHopAttr.NextHop != "" {
						break // Found the next hop, stop further processing
					}
				default:
					fmt.Printf("unknown attribute type %s for %s prefix\n", attr.GetTypeUrl(), dest.Prefix)
				}
			}

			// Parse the attributes into the Route structure
			route := Route{
				Prefix:       prefix.String(),
				PrefixLength: uint32(prefixLength),
				NextHop:      nextHopAttr.NextHop,
				Origin:       originAttr.Origin,
				Identifier:   dest.Paths[i].Identifier,
			}
			routes = append(routes, route)
		}
	}
	return routes, nil
}

func (g *GoBGPClient) UpdatePaths(ctx context.Context, routes []Route) error {
	// Set up the stream
	stream, err := g.client.AddPathStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open AddPathStream: %w", err)
	}

	paths := make([]*api.Path, len(routes))
	for i := range routes {
		// Marshal NLRI (route information) into *anypb.Any
		nlri, err := anypb.New(&api.IPAddressPrefix{
			Prefix:    routes[i].Prefix,
			PrefixLen: routes[i].PrefixLength,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal NLRI: %w", err)
		}

		// Marshal the attributes (Pattrs) into *anypb.Any
		originAttr, err := anypb.New(&api.OriginAttribute{
			Origin: routes[i].Origin,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal origin attribute: %w", err)
		}

		// Marshal the NextHop attribute into *anypb.Any
		nextHopAttr, err := anypb.New(&api.NextHopAttribute{
			NextHop: routes[i].NextHop,
		})
		if err != nil {
			return fmt.Errorf("failed to marshal next-hop attribute: %w", err)
		}

		// Construct each Path object
		paths[i] = &api.Path{
			Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
			Nlri:   nlri,
			Pattrs: []*anypb.Any{
				originAttr,
				nextHopAttr,
			},
			Identifier: routes[i].Identifier,
		}
	}

	// Construct the AddPathStreamRequest with multiple paths
	req := &api.AddPathStreamRequest{
		Paths: paths,
	}

	// Send the request through the gRPC stream
	if err := stream.Send(req); err != nil {
		return fmt.Errorf("failed to send path in AddPathStream: %w", err)
	}

	// Close the stream and receive the server's response
	_, err = stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to close AddPathStream: %w", err)
	}

	return nil
}

// DeletePath removes a specified BGP route (prefix) from GoBGP
func (g *GoBGPClient) DeletePath(ctx context.Context, route Route) error {
	// Marshal the NLRI (route information) into *anypb.Any
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    route.Prefix,
		PrefixLen: route.PrefixLength,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal NLRI for deletion: %w", err)
	}

	// Marshal the attributes (Pattrs) into *anypb.Any
	originAttr, err := anypb.New(&api.OriginAttribute{
		Origin: route.Origin,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal origin attribute: %w", err)
	}

	// Marshal the NextHop attribute into *anypb.Any (if required)
	nextHopAttr, err := anypb.New(&api.NextHopAttribute{
		NextHop: route.NextHop,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal next-hop attribute for deletion: %w", err)
	}

	// Construct the Path object with the NLRI and NextHop
	path := &api.Path{
		Nlri: nlri,
		Family: &api.Family{
			Afi:  api.Family_AFI_IP,
			Safi: api.Family_SAFI_UNICAST,
		},
		Pattrs: []*anypb.Any{
			originAttr,
			nextHopAttr,
		},
		Identifier: route.Identifier,
	}

	// Call DeletePath API with the constructed path
	_, err = g.client.DeletePath(ctx, &api.DeletePathRequest{
		Path: path,
	})
	if err != nil {
		return fmt.Errorf("failed to delete path from GoBGP: %w", err)
	}

	return nil
}
