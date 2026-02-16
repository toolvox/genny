// Package generator provides core domain types and generation logic for the genny static site generator.
// It includes types for Site, Component, Page, Asset, and DataContext, along with
// generators for creating component previews and main site pages.
package generator

import "html/template"

// Site represents the entire static site project with all its resources
type Site struct {
	RootPath   string
	Assets     []Asset
	Data       DataContext
	Components map[string]*Component
	Pages      []*Page
	Templates  map[string]*template.Template
}

// Component represents a reusable HTML component with its template and data requirements
type Component struct {
	Name         string
	FilePath     string
	Template     string
	DataPath     string
	Dependencies []string // Names of other components this component references
}

// Page represents a single output HTML page
type Page struct {
	SourcePath  string      // Source file path
	OutputPath  string      // Output file path (relative to www/)
	Content     string      // Raw HTML content
	Template    string      // Processed template content
	DataContext interface{} // Data for template execution
	IsPreview   bool        // True for component previews, false for main site pages
	EncryptKey  string      // If set, the page output will be encrypted with this passphrase
}

// Asset represents a static asset file (image, font, etc.)
type Asset struct {
	SourcePath string
	OutputPath string
}

// DataContext provides type-safe access to YAML data
type DataContext interface {
	// Get retrieves data at the given dot-separated path (e.g., "Posts.Featured")
	Get(path string) (interface{}, error)

	// GetAll returns all loaded data
	GetAll() map[string]interface{}

	// Set stores data at the given path
	Set(path string, value interface{}) error
}

// SimpleDataContext is a basic implementation of DataContext
type SimpleDataContext struct {
	data map[string]interface{}
}

// NewSimpleDataContext creates a new SimpleDataContext
func NewSimpleDataContext(data map[string]interface{}) *SimpleDataContext {
	if data == nil {
		data = make(map[string]interface{})
	}
	return &SimpleDataContext{data: data}
}

// Get retrieves data at the given dot-separated path
func (ctx *SimpleDataContext) Get(path string) (interface{}, error) {
	if path == "" || path == "." {
		return ctx.data, nil
	}

	parts := splitPath(path)
	var current interface{} = ctx.data

	for _, part := range parts {
		if part == "" {
			continue
		}

		currentMap, ok := current.(map[string]interface{})
		if !ok {
			return nil, &DataPathError{Path: path, Part: part}
		}

		current, ok = currentMap[part]
		if !ok {
			return nil, &DataPathError{Path: path, Part: part}
		}
	}

	return current, nil
}

// GetAll returns all loaded data
func (ctx *SimpleDataContext) GetAll() map[string]interface{} {
	return ctx.data
}

// Set stores data at the given path
func (ctx *SimpleDataContext) Set(path string, value interface{}) error {
	if path == "" || path == "." {
		if m, ok := value.(map[string]interface{}); ok {
			ctx.data = m
			return nil
		}
		return &DataPathError{Path: path, Part: "root"}
	}

	parts := splitPath(path)
	current := ctx.data

	for i, part := range parts[:len(parts)-1] {
		if part == "" {
			continue
		}

		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}

		nextMap, ok := current[part].(map[string]interface{})
		if !ok {
			return &DataPathError{Path: path, Part: parts[i]}
		}
		current = nextMap
	}

	lastPart := parts[len(parts)-1]
	current[lastPart] = value
	return nil
}

// splitPath splits a dot-separated path into parts
func splitPath(path string) []string {
	if path == "" || path == "." {
		return []string{}
	}

	var parts []string
	current := ""

	for _, char := range path {
		if char == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
