package model

// DatabaseAdapter defines interface for database communication
type DatabaseAdapter interface {
	HealthCheck() error
	Get(key string) (string, error)
	Put(key, value string) error
	Patch(key, value string) error
	Delete(key string) error
}
