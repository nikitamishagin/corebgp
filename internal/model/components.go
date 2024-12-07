package model

// Components represents a structure for tracking the state of various components with their enabled status.
type Components struct {
	Checker    bool `json:"checker"`
	IPAMPlugin bool `json:"ipam-plugin"`
}
