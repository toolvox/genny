// Package cli handles command-line interface argument parsing.
// It supports flags for watch mode, verbose logging, and specifying the project path.
package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
)

// Config holds CLI configuration
type Config struct {
	RootPath string
	Watch    bool
	Verbose  bool
}

// ParseArgs parses command line arguments
func ParseArgs() (*Config, error) {
	config := &Config{}

	// Define flags
	watch := flag.Bool("watch", false, "Watch for file changes and regenerate automatically")
	watchShort := flag.Bool("w", false, "Watch for file changes (shorthand)")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	verboseShort := flag.Bool("v", false, "Enable verbose logging (shorthand)")
	help := flag.Bool("help", false, "Show help message")
	helpShort := flag.Bool("h", false, "Show help message (shorthand)")

	flag.Parse()

	// Check for help flag first
	if *help || *helpShort {
		PrintUsage()
		os.Exit(0)
	}

	// Set watch mode (either -watch or -w)
	config.Watch = *watch || *watchShort
	log.Printf("watching: %t", config.Watch)

	// Set verbose mode (either -verbose or -v)
	config.Verbose = *verbose || *verboseShort
	log.Printf("verbose: %t", config.Verbose)

	// Get root path from positional argument or use current directory
	args := flag.Args()
	if len(args) > 0 {
		config.RootPath = args[0]
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		config.RootPath = wd
	}

	return config, nil
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Println("genny - Static site generator")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  genny [flags] [path]")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  path          Project directory (defaults to current directory)")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -w, -watch    Watch for file changes and regenerate automatically")
	fmt.Println("  -v, -verbose  Enable verbose logging")
	fmt.Println("  -h, -help     Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  genny                  # Generate site in current directory")
	fmt.Println("  genny ./mysite         # Generate site in ./mysite")
	fmt.Println("  genny -w               # Generate and watch for changes")
	fmt.Println("  genny -w -v ./mysite   # Generate, watch, and show verbose output")
}
