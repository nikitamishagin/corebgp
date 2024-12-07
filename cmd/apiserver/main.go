package main

import (
	"github.com/nikitamishagin/corebgp/internal/apiserver"
	"log"
)

func main() {
	err := apiserver.Run()
	if err != nil {
		log.Fatalf("Failed to run apiserver: %v", err)
	}
}
