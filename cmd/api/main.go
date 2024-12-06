package main

import (
	"github.com/nikitamishagin/corebgp/internal/api"
	"log"
)

func main() {
	if err := api.Run(); err != nil {
		log.Fatalf("Error starting apiserver: %v", err)
	}
}
