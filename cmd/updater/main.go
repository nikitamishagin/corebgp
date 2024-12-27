package main

import (
	"github.com/nikitamishagin/corebgp/internal/updater"
	"log"
)

// main is the entry point of the application that starts the CoreBGP updater.
func main() {
	err := updater.RootCmd().Execute()
	if err != nil {
		log.Fatalf("failed to run updater: %v", err)
	}
}
