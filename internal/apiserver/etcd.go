package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"go.etcd.io/etcd/client/v3"
	"os"
	"time"
)

// EtcdClient is a wrapper around the etcd client to simplify interactions and manage connection lifecycle.
type EtcdClient struct {
	client *clientv3.Client
}

// NewEtcdClient creates and initializes a new EtcdClient with the provided endpoints and TLS credentials.
func NewEtcdClient(endpoints []string, caFile, certFile, keyFile string) (*EtcdClient, error) {
	caCert, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("could not load client certificate and key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 3 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	return &EtcdClient{client: cli}, nil
}

// Close gracefully closes the underlying etcd client connection and releases associated resources.
func (e *EtcdClient) Close() {
	_ = e.client.Close()
}

// HealthCheck verifies the health status of the etcd client by querying the status of the first endpoint.
// It returns an error if the health check fails, indicating that the etcd client may be unreachable.
func (e *EtcdClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple status check via the shortest operation
	_, err := e.client.Status(ctx, e.client.Endpoints()[0])
	if err != nil {
		return fmt.Errorf("etcd health check failed: %w", err)
	}
	return nil
}

// Put inserts or updates a key-value pair in the etcd store.
func (e *EtcdClient) Put(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put data to etcd: %w", err)
	}
	return nil
}

// Get retrieves the value associated with the given key from etcd store.
func (e *EtcdClient) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get data from etcd: %w", err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found")
	}

	value := string(resp.Kvs[0].Value)
	return value, nil
}

// List retrieves all keys with the specified prefix from the etcd store and returns them as a slice of strings.
func (e *EtcdClient) List(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get data from etcd: %w", err)
	}

	keys := make([]string, len(resp.Kvs))
	for i := range resp.Kvs {
		keys[i] = string(resp.Kvs[i].Key)
	}
	return keys, nil
}

// GetObjects retrieves all values associated with keys that share the specified prefix from the etcd store.
func (e *EtcdClient) GetObjects(prefix string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := e.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get data from etcd: %w", err)
	}

	values := make([]string, len(resp.Kvs))
	for i := range resp.Kvs {
		values[i] = string(resp.Kvs[i].Value)
	}

	return values, nil
}

// Delete removes the key-value pair associated with the specified key from the etcd store.
func (e *EtcdClient) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete data from etcd: %w", err)
	}
	return nil
}

// Watch sets up a watch operation on a specified key and streams events through a channel until the stop signal is received.
// The stopChan is used to terminate the watch operation by canceling the associated context.
func (e *EtcdClient) Watch(key string, stopChan <-chan struct{}) (<-chan clientv3.WatchResponse, error) {
	// Create a context that can be canceled to stop the watch operation
	ctx, cancel := context.WithCancel(context.Background())

	// Start a goroutine to listen for a signal on stopChan
	go func() {
		select {
		case <-stopChan:
			// Stop the context when a signal is received on stopChan
			cancel()
		case <-ctx.Done():
			// If the context is already done, exit the goroutine
		}
	}()

	// Start watching the specified key with a prefix
	// The returned channel streams events; the caller is responsible for processing them
	return e.client.Watch(ctx, key, clientv3.WithPrefix(), clientv3.WithPrevKV()), nil
}
