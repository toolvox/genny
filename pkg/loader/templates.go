package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"genny/pkg/utils"
)

// LoadTemplates loads template files (index.html, header.html, footer.html)
func (l *FileSystemLoader) LoadTemplates(root string) (map[string]string, error) {
	templates := make(map[string]string)

	// Load index.html (required, keep full content)
	indexPath := filepath.Join(root, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("required template file not found: index.html: %w", err)
	}
	templates["index.html"] = string(indexContent)

	// Load header.html (optional, extract body content only)
	headerPath := filepath.Join(root, "header.html")
	if headerContent, err := utils.ExtractBodyContent(headerPath); err == nil {
		templates["header.html"] = headerContent
	}

	// Load footer.html (optional, extract body content only)
	footerPath := filepath.Join(root, "footer.html")
	if footerContent, err := utils.ExtractBodyContent(footerPath); err == nil {
		templates["footer.html"] = footerContent
	}

	return templates, nil
}
