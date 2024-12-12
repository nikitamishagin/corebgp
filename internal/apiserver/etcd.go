package apiserver

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/nikitamishagin/corebgp/internal/model"
	"go.etcd.io/etcd/client/v3"
	"os"
	"time"
)

type EtcdClient struct {
	client *clientv3.Client
}

func NewEtcdClient(config *model.APIConfig) (*EtcdClient, error) {
	caCert, err := os.ReadFile(config.EtcdCACert)
	if err != nil {
		return nil, fmt.Errorf("could not read CA certificate: %w", err)
	}
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	cert, err := tls.LoadX509KeyPair(config.EtcdClientCert, config.EtcdClientKey)
	if err != nil {
		return nil, fmt.Errorf("could not load client certificate and key: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.EtcdEndpoints,
		DialTimeout: 3 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	return &EtcdClient{client: cli}, nil
}

// CheckHealth verifies the health status of the etcd client by querying the status of the first endpoint.
// It returns an error if the health check fails, indicating that the etcd client may be unreachable.
func (e *EtcdClient) CheckHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Simple status check via the shortest operation
	_, err := e.client.Status(ctx, e.client.Endpoints()[0])
	if err != nil {
		return fmt.Errorf("etcd health check failed: %w", err)
	}
	return nil
}

func (e *EtcdClient) PutData(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put data to etcd: %w", err)
	}
	return nil
}

func (e *EtcdClient) GetData(key string) (string, error) {
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

func (e *EtcdClient) DeleteData(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := e.client.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete data from etcd: %w", err)
	}
	return nil
}
