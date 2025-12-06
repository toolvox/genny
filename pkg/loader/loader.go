// Package loader handles loading all project resources from the file system.
// It provides interfaces and implementations for loading assets, data files,
// components, and templates.
package loader

import "genny/pkg/generator"

// Loader handles loading all project resources
type Loader interface {
	// LoadAssets discovers and loads all static assets
	LoadAssets(root string) ([]generator.Asset, error)

	// LoadData loads and merges all YAML data files
	LoadData(root string) (map[string]interface{}, error)

	// LoadComponents discovers and loads all component files
	LoadComponents(root string) (map[string]*generator.Component, error)

	// LoadTemplates loads template files (index.html, header.html, footer.html)
	LoadTemplates(root string) (map[string]string, error)

	// LoadPages discovers and loads all page files from subdirectories
	LoadPages(root string) ([]*generator.Page, error)
}

// FileSystemLoader implements Loader using the file system
type FileSystemLoader struct{}

// NewFileSystemLoader creates a new FileSystemLoader
func NewFileSystemLoader() *FileSystemLoader {
	return &FileSystemLoader{}
}
