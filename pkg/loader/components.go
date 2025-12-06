package loader

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"genny/pkg/generator"
)

// LoadComponents discovers and loads all component files from the components directory
func (l *FileSystemLoader) LoadComponents(root string) (map[string]*generator.Component, error) {
	componentsPath := filepath.Join(root, "components")
	components := make(map[string]*generator.Component)

	err := filepath.WalkDir(componentsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If components directory doesn't exist, that's okay - just return empty map
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".html") {
			return nil
		}

		name := strings.TrimSuffix(d.Name(), ".html")

		// Check for duplicates
		if _, exists := components[name]; exists {
			return fmt.Errorf("duplicate component: %s", name)
		}

		components[name] = &generator.Component{
			Name:     name,
			FilePath: path,
			// Template and DataPath will be filled in by parser
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk components directory: %w", err)
	}

	return components, nil
}
