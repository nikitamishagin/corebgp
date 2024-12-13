package model

import clientv3 "go.etcd.io/etcd/client/v3"

// DatabaseAdapter defines interface for database communication
type DatabaseAdapter interface {
	HealthCheck() error
	Get(key string) (string, error)
	Put(key, value string) error
	Patch(key, value string) error
	Watch(key string, stopChan <-chan struct{}) (<-chan clientv3.WatchResponse, error)
	Delete(key string) error
}
