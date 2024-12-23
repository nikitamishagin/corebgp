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
