package loader

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/toolvox/utilgo/pkg/serialization/yaml"
)

// LoadData loads and merges all YAML data files from the data directory
func (l *FileSystemLoader) LoadData(root string) (map[string]interface{}, error) {
	dataPath := filepath.Join(root, "data")
	result := make(map[string]interface{})

	err := filepath.WalkDir(dataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If data directory doesn't exist, that's okay - just return empty map
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
			return nil
		}

		data, err := yaml.UnmarshalFile[map[string]interface{}](path)
		if err != nil {
			return fmt.Errorf("failed to parse YAML file %s: %w", path, err)
		}

		filename := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
		result[filename] = data

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk data directory: %w", err)
	}

	return result, nil
}
