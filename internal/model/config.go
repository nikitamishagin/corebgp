package model

// APIConfig represents the configuration parameters required to initialize and run the API server.
type APIConfig struct {
	DBType    string   `yaml:"db_type"`   // DBType specifies the type of database to be used, e.g., "etcd".
	Endpoints []string `yaml:"endpoints"` // Endpoints defines the list of database endpoint URLs for connecting the API server to the database backend.
	Etcd      Etcd     `yaml:"etcd"`      // Etcd contains the configuration details needed to connect to an Etcd cluster.
	TLSCert   string   `yaml:"tls_cert"`  // TLSCert specifies the file path to the TLS certificate used for securing API server communication.
	TLSKey    string   `yaml:"tls_key"`   // TLSKey specifies the file path to the TLS private key used for securing API server communication.
	LogPath   string   `yaml:"log_path"`  // LogPath specifies the file path to the log file for storing API server logs.
	Verbose   int8     `yaml:"verbose"`   // Verbose specifies the verbosity level for logging, where higher values produce more detailed logs.
}

// Etcd is a configuration structure used for specifying Etcd cluster connection parameters.
type Etcd struct {
	CACert     string `yaml:"ca_cert"`     // CACert specifies the file path to the CA certificate to establish secure communication with the Etcd cluster.
	ClientCert string `yaml:"client_cert"` // ClientCert specifies the file path to the client certificate for authenticating with the Etcd cluster.
	ClientKey  string `yaml:"client_key"`  // ClientKey specifies the file path to the client private key used for authenticating with the Etcd cluster.
}

// UpdaterConfig represents the configuration parameters required to initialize and run the Updater controller.
type UpdaterConfig struct {
	APIEndpoint     string `yaml:"api_endpoint"`      // APIEndpoint specifies the URL to the API server endpoint.
	GoBGPEndpoint   string `yaml:"gobgp_endpoint"`    // GoBGPEndpoint specifies the URL to the GoBGP API.
	GoBGPCACert     string `yaml:"gobgp_ca_cert"`     // GoBGPCACert specifies the path to the GoBGP CA certificate file.
	GoBGPClientCert string `yaml:"gobgp_client_cert"` // GoBGPClientCert specifies the path to the GoBGP client certificate file.
	GoBGPClientKey  string `yaml:"gobgp_client_key"`  // GoBGPClientKey specifies the path to the GoBGP client key file.
	LogPath         string `yaml:"log_path"`          // LogPath specifies the file path to the log file for storing updater logs.
	Verbose         int8   `yaml:"verbose"`           // Verbose specifies the verbosity level for logging, where higher values produce more detailed logs.
}

// CheckerConfig represents the configuration parameters required to initialize and run the Checker system.
type CheckerConfig struct {
	APIEndpoint     string `yaml:"api_endpoint"`     // APIEndpoint specifies the URL to the API server endpoint.
	Zone            string `yaml:"zone"`             // Zone specifies the geographic or logical zone where the Checker system is deployed.
	LivenessTimeout string `yaml:"liveness_timeout"` // LivenessTimeout specifies the time interval during which the component should update its health status.
	LogPath         string `yaml:"log_path"`         // LogPath specifies the file path to the log file for storing checker logs.
	Verbose         int8   `yaml:"verbose"`          // Verbose specifies the verbosity level for logging, where higher values produce more detailed logs.
}
