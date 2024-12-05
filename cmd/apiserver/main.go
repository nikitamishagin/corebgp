package main

import (
	"github.com/nikitamishagin/corebgp/internal/apiserver"
	"github.com/nikitamishagin/corebgp/internal/etcd"
	"log"

	"github.com/spf13/cobra"
)

var config string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "apiserver",
		Short: "API Server is a RESTful server for interacting with etcd",
		Run: func(cmd *cobra.Command, args []string) {
			client, err := etcd.NewClient(config)
			if err != nil {
				log.Fatalf("Failed to connect to etcd: %v", err)
			}

			server := apiserver.NewAPIServer(client)
			server.Start()
		},
	}

	rootCmd.Flags().StringVarP(&config, "config", "c", "config/apiserver-config.yaml", "Configuration file")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}
}
