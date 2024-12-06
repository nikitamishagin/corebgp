package model

// Announce represents the configuration and health status of a system component or service.
type Announce struct {
	Meta         Meta         `json:"meta"`
	Addresses    Addresses    `json:"addresses"`
	Endpoints    Endpoints    `json:"endpoints"`
	HealthChecks HealthChecks `json:"health-checks"`
	Status       Status       `json:"status"`
}

// Meta contains metadata information about bgp announce such as version, name, uuid, and project.
type Meta struct {
	Version string `json:"version"`
	Name    string `json:"name"`
	UUID    string `json:"UUID"`
	Project string `json:"project"`
}

// Addresses represents a collection of network-related data, including subnets, zone, and announcing ips.
type Addresses struct {
	Subnets IPAddress   `json:"subnets"`
	Zone    string      `json:"zone"`
	IPs     []IPAddress `json:"IPs"`
}

type IPAddress struct {
	IP   string `json:"ip"`
	Mask uint8  `json:"mask"`
}

type Endpoints struct {
	Hosts []string `json:"hosts"`
}

type HealthChecks struct {
	Path               string `json:"path"`
	Port               int    `json:"port"`
	Method             string `json:"method"`
	IntervalSeconds    int    `json:"interval-seconds"`
	TimeoutSeconds     int    `json:"timeout-seconds"`
	GracePeriodSeconds int    `json:"grace-period-seconds"`
}

type Status struct {
	Status    string  `json:"status"`
	Details   Details `json:"details"`
	Timestamp string  `json:"timestamp"`
}

type Details struct {
	Host      string `json:"host"`
	Status    string `json:"status"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	Timestamp string `json:"timestamp"`
}
