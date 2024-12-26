package updater

import (
	"context"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/anypb"
	"time"

	api "github.com/osrg/gobgp/v3/api"
)

// GoBGPClient is struct for manage GoBGP client
type GoBGPClient struct {
	client api.GobgpApiClient
	conn   *grpc.ClientConn
}

// NewGoBGPClient initializes the new GoBGP client
func NewGoBGPClient(config *model.UpdaterConfig) (*GoBGPClient, error) {
	conn, err := grpc.Dial(config.GoBGPEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	client := api.NewGobgpApiClient(conn)
	return &GoBGPClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes GoBGP API server connection
func (g *GoBGPClient) Close() {
	if g.conn != nil {
		g.conn.Close()
	}
}

// AddPath adds a specified BGP route (prefix) with associated attributes to the GoBGP server.
func (g *GoBGPClient) AddPath(prefix string, prefixLength uint32, nextHop string) error {
	// Generate the context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Marshal the NLRI (route information) into *anypb.Any
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
		return fmt.Errorf("failed to marshal NLRI for deletion: %w", err)
	}

	// Marshal the NextHop attribute into *anypb.Any (if required)
	nextHopAttr, err := anypb.New(&api.NextHopAttribute{
		NextHop: nextHop,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal next-hop attribute for deletion: %w", err)
	}

	// Construct the Path object
	path := &api.Path{
		Family: &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST},
		Nlri:   nlri,
		Pattrs: []*anypb.Any{
			originAttr,
			nextHopAttr,
		},
	}

	// Add the route to the GoBGP server
	_, err = g.client.AddPath(ctx, &api.AddPathRequest{
		Path: path,
	})
	if err != nil {
		return fmt.Errorf("failed to add path to GoBGP: %w", err)
	}

	return nil
}

// DeletePath removes a specified BGP route (prefix) from GoBGP
func (g *GoBGPClient) DeletePath(prefix string, prefixLength uint32, nextHop string) error {
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

	// Marshal the NextHop attribute into *anypb.Any (if required)
	nextHopAttr, err := anypb.New(&api.NextHopAttribute{
		NextHop: nextHop,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal next-hop attribute for deletion: %w", err)
	}

	// Construct the Path object with the NLRI and NextHop
	path := &api.Path{
		Nlri: nlri,
		Pattrs: []*anypb.Any{
			nextHopAttr,
		},
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
