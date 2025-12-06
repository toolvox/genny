// Package orchestrator coordinates the site generation workflow.
// It provides RunOnce for single generation and RunContinuous for watch mode.
package orchestrator

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"genny/pkg/site"
	"genny/pkg/watcher"
)

// Orchestrator coordinates the site generation workflow
type Orchestrator struct {
	site    *site.Site
	watcher watcher.Watcher
	verbose bool
}

// NewOrchestrator creates a new Orchestrator
func NewOrchestrator(rootPath string, verbose bool) *Orchestrator {
	return &Orchestrator{
		site:    site.NewSite(rootPath, verbose),
		watcher: watcher.NewFileWatcher(500 * time.Millisecond),
		verbose: verbose,
	}
}

// RunOnce loads and generates the site once
func (o *Orchestrator) RunOnce() error {
	start := time.Now()

	log.Println("Starting site generation...")

	// Load the site
	if err := o.site.Load(); err != nil {
		return fmt.Errorf("failed to load site: %w", err)
	}

	// Generate the site
	if err := o.site.Generate(); err != nil {
		return fmt.Errorf("failed to generate site: %w", err)
	}

	elapsed := time.Since(start)
	log.Printf("✓ Site generated successfully in %v", elapsed)

	return nil
}

// RunContinuous runs in watch mode, regenerating on file changes
func (o *Orchestrator) RunContinuous() error {
	// Initial generation
	if err := o.RunOnce(); err != nil {
		return err
	}

	log.Println()
	log.Println("Watching for changes... (Press Ctrl+C to stop)")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Watch paths - directories
	watchPaths := []string{
		".", // Watch root directory for HTML/CSS files
		"./data",
		"./components",
		"./assets",
	}

	// Add all page files from subdirectories
	if o.site.GetSite() != nil {
		for _, page := range o.site.GetSite().Pages {
			watchPaths = append(watchPaths, page.SourcePath)
		}
	}

	// Create a channel for regeneration
	regenerateChan := make(chan string, 10)

	// Start the watcher in a goroutine
	watcherErrChan := make(chan error, 1)
	go func() {
		err := o.watcher.Watch(watchPaths, func(path string) {
			regenerateChan <- path
		})
		watcherErrChan <- err
	}()

	// Give the watcher a moment to start
	time.Sleep(100 * time.Millisecond)

	// Main loop
	for {
		select {
		case <-sigChan:
			log.Println()
			log.Println("Shutting down gracefully...")
			if err := o.watcher.Stop(); err != nil {
				log.Printf("Warning: Error stopping watcher: %v", err)
			}
			return nil

		case err := <-watcherErrChan:
			if err != nil {
				log.Printf("Error: Watch mode failed: %v", err)
				log.Println("Exiting...")
				return err
			}
			log.Println("Watcher stopped")
			return nil

		case path := <-regenerateChan:
			timestamp := time.Now().Format("15:04:05")
			log.Printf("[%s] Changed: %s → regenerating...", timestamp, path)

			start := time.Now()
			if err := o.RunOnce(); err != nil {
				log.Printf("✗ Regeneration failed: %v", err)
			} else {
				elapsed := time.Since(start)
				log.Printf("[%s] ✓ Regenerated in %v", time.Now().Format("15:04:05"), elapsed)
			}
		}
	}
}
