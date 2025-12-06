package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"genny/pkg/generator"
)

// LoadPages discovers all .html files at root level (excluding index.html, header.html, footer.html)
// and index.html files in subdirectories (excluding components, data, assets, www)
func (l *FileSystemLoader) LoadPages(root string) ([]*generator.Page, error) {
	var pages []*generator.Page

	// Walk through all directories
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip if not a file
		if info.IsDir() {
			return nil
		}

		// Only process .html files
		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		// Get relative path from root
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		// Skip special files at root level
		if relPath == "index.html" || relPath == "header.html" || relPath == "footer.html" {
			return nil
		}

		// Check if file is at root level (no subdirectory)
		dir := filepath.Dir(relPath)
		isRootLevel := dir == "."

		if isRootLevel {
			// Accept any .html file at root level (except index, header, footer)
			// Output path stays the same
		} else {
			// For subdirectories: only process index.html files
			if info.Name() != "index.html" {
				return nil
			}

			// Skip special directories
			if strings.HasPrefix(dir, "components") ||
				strings.HasPrefix(dir, "data") ||
				strings.HasPrefix(dir, "assets") ||
				strings.HasPrefix(dir, "www") {
				return nil
			}
		}

		// Read the page content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read page %s: %w", path, err)
		}

		// Create Page struct
		page := &generator.Page{
			SourcePath: path,
			OutputPath: relPath,
			Content:    string(content),
		}

		pages = append(pages, page)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover pages: %w", err)
	}

	return pages, nil
}
