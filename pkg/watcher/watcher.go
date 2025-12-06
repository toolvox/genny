// Package watcher provides file system monitoring for automatic regeneration.
// It uses fsnotify for cross-platform file watching with debouncing support.
package watcher

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher handles file system watching for changes
type Watcher interface {
	// Watch starts watching the specified paths and calls onChange when changes are detected
	Watch(paths []string, onChange func(path string)) error

	// Stop stops the watcher
	Stop() error
}

// FileWatcher implements Watcher using fsnotify
type FileWatcher struct {
	debounceInterval time.Duration
	stopChan         chan bool
	watcher          *fsnotify.Watcher
	debounceTimer    *time.Timer
	pendingChanges   map[string]bool
}

// NewFileWatcher creates a new FileWatcher
func NewFileWatcher(debounceInterval time.Duration) *FileWatcher {
	return &FileWatcher{
		debounceInterval: debounceInterval,
		stopChan:         make(chan bool),
		pendingChanges:   make(map[string]bool),
	}
}

// Watch starts watching the specified paths and calls onChange when changes are detected
func (w *FileWatcher) Watch(paths []string, onChange func(path string)) error {
	var err error
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.watcher.Close()

	// Add paths to watcher
	for _, path := range paths {
		// Check if path exists
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				log.Printf("Warning: Watch path does not exist: %s (skipping)", path)
				continue
			}
			return err
		}

		// Add to watcher
		if err := w.watcher.Add(path); err != nil {
			log.Printf("Warning: Could not watch %s: %v", path, err)
			continue
		}

		// If it's a directory, also watch files in it (non-recursive for now)
		info, _ := os.Stat(path)
		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				log.Printf("Warning: Could not read directory %s: %v", path, err)
				continue
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					fullPath := filepath.Join(path, entry.Name())
					if err := w.watcher.Add(fullPath); err != nil {
						log.Printf("Warning: Could not watch %s: %v", fullPath, err)
					}
				}
			}
		}
	}

	// Main watch loop
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			// Only process Write and Create events
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				// Add to pending changes
				w.pendingChanges[event.Name] = true

				// Reset or create debounce timer
				if w.debounceTimer != nil {
					w.debounceTimer.Stop()
				}

				w.debounceTimer = time.AfterFunc(w.debounceInterval, func() {
					// Process all pending changes
					for path := range w.pendingChanges {
						onChange(path)
					}
					// Clear pending changes
					w.pendingChanges = make(map[string]bool)
				})
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watcher error: %v", err)

		case <-w.stopChan:
			if w.debounceTimer != nil {
				w.debounceTimer.Stop()
			}
			return nil
		}
	}
}

// Stop stops the watcher
func (w *FileWatcher) Stop() error {
	close(w.stopChan)
	return nil
}
