package checker

import (
	"github.com/nikitamishagin/corebgp/internal/checker"
	"log"
)

// main is the entry point of the application that starts the CoreBGP checker.
func main() {
	err := checker.RootCmd().Execute()
	if err != nil {
		log.Fatalf("failed to run checker: %v", err)
	}
}
