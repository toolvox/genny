package main

import (
	"log"
	"os"

	"genny/pkg/cli"
	"genny/pkg/orchestrator"

	"github.com/toolvox/utilgo/pkg/errs"
)

const version = "v0.1.1"

func main() {
	log.Printf("Genny %s", version)

	// Parse command line arguments
	config, err := cli.ParseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %v", err)
	}

	log.Printf("starting directory: %s", errs.Must(os.Getwd()))
	// Change to the specified directory
	if err := os.Chdir(config.RootPath); err != nil {
		log.Fatalf("Error changing to directory %s: %v", config.RootPath, err)
	}

	// Create orchestrator
	orch := orchestrator.NewOrchestrator(".", config.Verbose)

	// Run in appropriate mode
	if config.Watch {
		if err := orch.RunContinuous(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	} else {
		if err := orch.RunOnce(); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}
}
