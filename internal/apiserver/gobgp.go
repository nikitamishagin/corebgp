package apiserver

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
func NewGoBGPClient(config *model.APIConfig) (*GoBGPClient, error) {
	conn, err := grpc.Dial(config.GoBGPInstance, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// TODO: AddAnnouncement needs to be completed

// AddAnnouncement adds a BGP announcement based on the given Announcement structure.
func (g *GoBGPClient) AddAnnouncement(announcement model.Announcement) error {
	// Generate the context for the gRPC call
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Validate the required fields before proceeding
	if announcement.Addresses.AnnouncedIP == "" && announcement.Addresses.SourceSubnets.IP == "" {
		return fmt.Errorf("necessarily one of the following fields must be set: announced_ip, source_subnets")
	}
	if len(announcement.NextHops) == 0 {
		return fmt.Errorf("next_hops must be set")
	}

	// Marshal the IP prefix (NLRI) into *anypb.Any
	nlri, err := anypb.New(&api.IPAddressPrefix{
		Prefix:    announcement.Addresses.AnnouncedIP,
		PrefixLen: 32,
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

	nextHopAttr, err := anypb.New(&api.NextHopAttribute{
		NextHop: announcement.NextHops[0].IP, // Use the first next-hop from the list
	})
	if err != nil {
		return fmt.Errorf("failed to marshal next-hop attribute: %w", err)
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
		return fmt.Errorf("failed to add announcement: %w", err)
	}

	return nil
}
