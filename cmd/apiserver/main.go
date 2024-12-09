package main

import (
	"github.com/nikitamishagin/corebgp/internal/apiserver"
	"log"
)

// main is the entry point of the application that starts the CoreBGP API server.
func main() {
	err := apiserver.RootCmd().Execute()
	if err != nil {
		log.Fatalf("Failed to run apiserver: %v", err)
	}
}
