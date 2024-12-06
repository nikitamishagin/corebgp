package etcd

import (
	"context"
	"log"
	"time"

	etcdClientV3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	cli *etcdClientV3.Client
}

func NewClient() (*Client, error) {
	// This should read the configuration from configPath (e.g., YAML, JSON, etc.).
	// Here we're just simulating the client creation.
	cfg := etcdClientV3.Config{
		Endpoints: []string{"localhost:2379"}, // Replace with endpoints from config
		//DialTimeout: 5 * etcdClientV3.DefaultDialTimeout,
	}

	cli, err := etcdClientV3.New(cfg)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to etcd")
	return &Client{cli: cli}, nil
}

func (c *Client) Write(key, value string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_, err := c.cli.Put(ctx, key, value)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Read(key string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.cli.Get(ctx, key)
	if err != nil {
		log.Printf("Error getting key %s: %s", key, err)
		return "", err
	}

	if len(resp.Kvs) == 0 {
		log.Printf("Key %s not found", key)
		return "", nil // Key not found
	}

	return string(resp.Kvs[0].Value), nil
}

func (c *Client) Watch(key string, stopChan <-chan struct{}) {
	rch := c.cli.Watch(context.Background(), key)

	go func() {
		for {
			select {
			case resp := <-rch:
				for _, ev := range resp.Events {
					switch ev.Type {
					case etcdClientV3.EventTypePut:
						log.Printf("PUT Key: %s, Value: %s", ev.Kv.Key, ev.Kv.Value)
					case etcdClientV3.EventTypeDelete:
						log.Printf("DELETE Key: %s", ev.Kv.Key)
					}
				}
			case <-stopChan:
				log.Printf("Stopping watch on key %s", key)
				return
			}
		}
	}()
}
