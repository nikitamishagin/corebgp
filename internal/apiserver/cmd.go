package apiserver

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	//var config model.APIConfig
	var cmd = &cobra.Command{
		Use:   "apiserver",
		Short: "CoreBGP API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := NewAPIServer(); err != nil {
				return err
			}
			return nil
		},
	}

	//cmd.Flags().StringVar(&config.EtcdEndpoints, "etcd-endpoints", "http://localhost:2379", "Comma separated list of etcd endpoints")
	//cmd.Flags().StringVar(&config.EtcdCert, "etcd-cert", "", "Path to etcd client certificate")
	//cmd.Flags().StringVar(&config.EtcdKey, "etcd-key", "", "Path to etcd client key")
	//cmd.Flags().StringVar(&config.TlsCert, "tls-cert", "", "Path to TLS certificate")
	//cmd.Flags().StringVar(&config.TlsKey, "tls-key", "", "Path to TLS key")
	//cmd.Flags().StringVar(&config.GoBGPInstance, "gobgp-instance", "http://localhost:50051", "Endpoint of GoBGP instance")
	//cmd.Flags().StringVarP(&config.LogPath, "log-path", "l", "/var/log/corebgp/apiserver.log", "Path to log file")
	//cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}
