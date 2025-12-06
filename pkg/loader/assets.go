package loader

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"genny/pkg/generator"
)

// LoadAssets discovers and loads all static assets from the assets directory
func (l *FileSystemLoader) LoadAssets(root string) ([]generator.Asset, error) {
	assetsPath := filepath.Join(root, "assets")
	var assets []generator.Asset

	err := filepath.WalkDir(assetsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If assets directory doesn't exist, that's okay - just return empty list
			return nil
		}
		if d.IsDir() {
			return nil
		}

		// Convert absolute path to relative path from assets directory
		relPath, err := filepath.Rel(assetsPath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		assets = append(assets, generator.Asset{
			SourcePath: path,
			OutputPath: filepath.Join("assets", relPath),
		})
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk assets directory: %w", err)
	}

	return assets, nil
}
