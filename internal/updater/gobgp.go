package updater

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/anypb"
	"os"
	"time"

	api "github.com/osrg/gobgp/v3/api"
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
func (g *GoBGPClient) AddPaths(prefix string, prefixLength uint32, nextHops []string) error {
	// Generate the context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up the stream
	stream, err := g.client.AddPathStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open AddPathStream: %w", err)
	}

	// Marshal NLRI (route information) into *anypb.Any
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLength,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal NLRI: %w", err)
	}

	// Marshal the attributes (Pattrs) into *anypb.Any
	originAttr, err := anypb.New(&api.OriginAttribute{
		Origin: 0, // IGP
	})
	if err != nil {
		return fmt.Errorf("failed to marshal origin attribute: %w", err)
	}

	// Prepare a list of paths, one for each next-hop
	paths := make([]*api.Path, len(nextHops))
	for i := range nextHops {
		// Marshal the NextHop attribute into *anypb.Any
		nextHopAttr, err := anypb.New(&api.NextHopAttribute{
			NextHop: nextHops[i],
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
			Identifier: uint32(i + 1),
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

// ListPath retrieves a list of BGP paths for the specified prefix from the GoBGP server. Returns a slice of paths or an error.
func (g *GoBGPClient) ListPath(prefix string) ([]string, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call ListPath API with a prefix filter
	stream, err := g.client.ListPath(ctx, &api.ListPathRequest{
		Family: &api.Family{
			Afi:  api.Family_AFI_IP,
			Safi: api.Family_SAFI_UNICAST,
		},
		Prefixes: []*api.TableLookupPrefix{
			{
				Prefix: prefix,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list paths from GoBGP: %w", err)
	}

	// Collect paths from the stream
	var paths []string
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("error while receiving path from stream: %w", err)
		}
		paths = append(paths, resp.String())
	}

	return paths, nil
}

func (g *GoBGPClient) UpdatePath(prefix string, prefixLength uint32, nextHops []string) error {
	// Generate the context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up the stream
	stream, err := g.client.AddPathStream(ctx)
	if err != nil {
		return fmt.Errorf("failed to open AddPathStream: %w", err)
	}

	// Marshal NLRI (route information) into *anypb.Any
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLength,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal NLRI: %w", err)
	}

	// Marshal the attributes (Pattrs) into *anypb.Any
	originAttr, err := anypb.New(&api.OriginAttribute{
		Origin: 0, // IGP
	})
	if err != nil {
		return fmt.Errorf("failed to marshal origin attribute: %w", err)
	}

	// Prepare a list of paths, one for each next-hop
	paths := make([]*api.Path, len(nextHops))
	for i := range nextHops {
		// Marshal the NextHop attribute into *anypb.Any
		nextHopAttr, err := anypb.New(&api.NextHopAttribute{
			NextHop: nextHops[i],
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
			Identifier: uint32(i + 1),
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
func (g *GoBGPClient) DeletePath(prefix string, prefixLength uint32, nextHops []string) error {
	// Create context with timeout for gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Marshal the NLRI (route information) into *anypb.Any
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLength,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal NLRI for deletion: %w", err)
	}

	for i := range nextHops {
		// Marshal the NextHop attribute into *anypb.Any (if required)
		nextHopAttr, err := anypb.New(&api.NextHopAttribute{
			NextHop: nextHops[i],
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
				nextHopAttr,
			},
			Identifier: uint32(i + 1),
		}

		// Call DeletePath API with the constructed path
		_, err = g.client.DeletePath(ctx, &api.DeletePathRequest{
			Path: path,
		})
		if err != nil {
			return fmt.Errorf("failed to delete path from GoBGP: %w", err)
		}
	}

	return nil
}
