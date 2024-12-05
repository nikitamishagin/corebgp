package etcd

import (
	"log"

	etcdclientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	cli *etcdclientv3.Client
}

func NewClient(configPath string) (*Client, error) {
	// This should read the configuration from configPath (e.g., YAML, JSON, etc.).
	// Here we're just simulating the client creation.
	cfg := etcdclientv3.Config{
		Endpoints:   []string{"localhost:2379"}, // Replace with endpoints from config
		DialTimeout: 5 * etcdclientv3.DefaultDialTimeout,
	}

	cli, err := etcdclientv3.New(cfg)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to etcd")
	return &Client{cli: cli}, nil
}
