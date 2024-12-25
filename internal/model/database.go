package model

import clientv3 "go.etcd.io/etcd/client/v3"

// DatabaseAdapter defines interface for database communication
type DatabaseAdapter interface {
	HealthCheck() error
	Get(string) (string, error)
	List(string) ([]string, error)
	GetObjects(string) ([]string, error)
	Put(string, string) error
	Patch(string, string) error
	Watch(string, <-chan struct{}) (<-chan clientv3.WatchResponse, error)
	Delete(string) error
}
