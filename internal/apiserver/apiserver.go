package apiserver

// Run initializes and executes the root command for the CoreBGP API server application.
// Returns an error if execution fails.
func Run() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}
