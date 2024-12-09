package model

type APIConfig struct {
	EtcdEndpoints  string
	EtcdCACert     string
	EtcdClientCert string
	EtcdClientKey  string
	TlsCert        string
	TlsKey         string
	GoBGPInstance  string
	LogPath        string
	Verbose        int8
}
