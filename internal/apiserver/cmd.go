package apiserver

import (
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/spf13/cobra"
)

// RootCmd initializes and returns the root command for the CoreBGP API server application.
func RootCmd() *cobra.Command {
	var config model.APIConfig
	var cmd = &cobra.Command{
		Use:   "apiserver",
		Short: "CoreBGP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			//&config.EtcdEndpoints := strings.Split(etcdEndpoints, ",")
			//
			//// Checking that all elements in a list are not empty
			//for i, endpoint := range &config.EtcdEndpoints {
			//	endpoint = strings.TrimSpace(endpoint) // Remove spaces
			//	if endpoint == "" {
			//		return fmt.Errorf("etcd endpoint cannot be empty")
			//	}
			//	&config.EtcdEndpoints[i] := endpoint
			//}

			etcdClient, err := NewEtcdClient(&config)
			if err != nil {
				return err
			}
			if err := NewAPIServer(etcdClient); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&config.EtcdEndpoints, "etcd-endpoints", "http://localhost:2379", "Comma separated list of etcd endpoints")
	cmd.Flags().StringVar(&config.EtcdCACert, "etcd-ca", "", "Path to etcd CA certificate")
	cmd.Flags().StringVar(&config.EtcdClientCert, "etcd-cert", "", "Path to etcd client certificate")
	cmd.Flags().StringVar(&config.EtcdClientKey, "etcd-key", "", "Path to etcd client key")
	cmd.Flags().StringVar(&config.TlsCert, "tls-cert", "", "Path to TLS certificate")
	cmd.Flags().StringVar(&config.TlsKey, "tls-key", "", "Path to TLS key")
	cmd.Flags().StringVar(&config.GoBGPInstance, "gobgp-instance", "http://localhost:50051", "Endpoint of GoBGP instance")
	cmd.Flags().StringVarP(&config.LogPath, "log-path", "l", "/var/log/corebgp/apiserver.log", "Path to log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}
