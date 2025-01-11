package checker

import (
	"github.com/nikitamishagin/corebgp/internal/model"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	var config model.CheckerConfig
	var cmd = &cobra.Command{
		Use:   "checker",
		Short: "CoreBGP checker",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement checker logic
			return nil
		},
	}

	cmd.Flags().StringVar(&config.APIEndpoint, "api-endpoint", "http://localhost:8080", "URL of the API server")
	cmd.Flags().StringVar(&config.LogPath, "log-path", "/var/log/corebgp/checker.log", "Path to the log file")
	cmd.Flags().Int8VarP(&config.Verbose, "verbose", "v", 0, "Verbosity level")

	return cmd
}
