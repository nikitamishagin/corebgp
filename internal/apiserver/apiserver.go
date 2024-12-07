package apiserver

func Run() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}
