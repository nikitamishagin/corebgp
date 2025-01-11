package model

// Components represents a structure for tracking the state of various components with their enabled status.
type Components struct {
	Checker    bool `json:"checker"`
	IPAMPlugin bool `json:"ipam-plugin"`
}

// CheckerHealth represents the health status of the Checker component responsible for a specific zone.
type CheckerHealth struct {
	Zone             string `json:"zone"`               // Zone represents the specific geographical zone or region associated with the health check.
	LastResponseTime string `json:"last-response-time"` // LastResponseTime indicates the timestamp of the most recent response received from the health check.
}
