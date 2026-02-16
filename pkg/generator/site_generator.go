package generator

import (
	"bytes"
	"fmt"
	"genny/pkg/utils"
	"html/template"
	"os"
	"path/filepath"
)

// MainSiteGenerator handles generating the main site pages
type MainSiteGenerator struct {
	outputDir string
}

// NewMainSiteGenerator creates a new MainSiteGenerator
func NewMainSiteGenerator(outputDir string) *MainSiteGenerator {
	return &MainSiteGenerator{outputDir: outputDir}
}

// GenerateMainSite generates the main site using the index template
func (g *MainSiteGenerator) GenerateMainSite(site *Site, mainTemplateContent string, headerContent, footerContent string) error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a template set with all components, header, and footer
	t := template.New("Main")

	// Parse the main template first
	_, err := t.Parse(mainTemplateContent)
	if err != nil {
		return &TemplateParseError{
			Name:   "Main",
			Source: mainTemplateContent,
			Err:    err,
		}
	}

	// Add components
	for name, comp := range site.Components {
		_, err := t.New(name).Parse(comp.Template)
		if err != nil {
			return &TemplateParseError{
				Name:   name,
				Source: comp.Template,
				Err:    err,
			}
		}
	}

	// Add header and footer if they exist
	if headerContent != "" {
		_, err := t.New("header.html").Parse(headerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "header.html",
				Source: headerContent,
				Err:    err,
			}
		}
	}

	if footerContent != "" {
		_, err := t.New("footer.html").Parse(footerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "footer.html",
				Source: footerContent,
				Err:    err,
			}
		}
	}

	// Execute the main template with all data
	var buf bytes.Buffer
	if err := t.Execute(&buf, site.Data.GetAll()); err != nil {
		return &TemplateExecuteError{
			Name: "Main",
			Err:  err,
		}
	}

	// Clean up excessive whitespace
	cleaned := utils.CleanupWhitespace(buf.String())

	// Write to index.html in output directory
	outputPath := filepath.Join(g.outputDir, "index.html")
	if err := os.WriteFile(outputPath, []byte(cleaned), 0644); err != nil {
		return fmt.Errorf("failed to write main site file: %w", err)
	}

	return nil
}

// GeneratePages generates all pages from subdirectories
func (g *MainSiteGenerator) GeneratePages(site *Site, headerContent, footerContent string) error {
	for _, page := range site.Pages {
		if err := g.generatePage(page, site, headerContent, footerContent); err != nil {
			return fmt.Errorf("failed to generate page %s: %w", page.OutputPath, err)
		}
	}
	return nil
}

// generatePage generates a single page
func (g *MainSiteGenerator) generatePage(page *Page, site *Site, headerContent, footerContent string) error {
	// Create a template set with all components, header, and footer
	t := template.New("Page")

	// Parse the page content
	_, err := t.Parse(page.Content)
	if err != nil {
		return &TemplateParseError{
			Name:   page.OutputPath,
			Source: page.Content,
			Err:    err,
		}
	}

	// Add components
	for name, comp := range site.Components {
		_, err := t.New(name).Parse(comp.Template)
		if err != nil {
			return &TemplateParseError{
				Name:   name,
				Source: comp.Template,
				Err:    err,
			}
		}
	}

	// Add header and footer if they exist
	if headerContent != "" {
		_, err := t.New("header.html").Parse(headerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "header.html",
				Source: headerContent,
				Err:    err,
			}
		}
	}

	if footerContent != "" {
		_, err := t.New("footer.html").Parse(footerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "footer.html",
				Source: footerContent,
				Err:    err,
			}
		}
	}

	// Execute the page template with all data
	var buf bytes.Buffer
	if err := t.Execute(&buf, site.Data.GetAll()); err != nil {
		return &TemplateExecuteError{
			Name: page.OutputPath,
			Err:  err,
		}
	}

	// Clean up excessive whitespace
	cleaned := utils.CleanupWhitespace(buf.String())

	// Adjust paths based on directory depth
	// Calculate depth by counting path separators in the output path (excluding the filename)
	dir := filepath.Dir(page.OutputPath)
	depth := 0
	if dir != "." {
		depth = len(filepath.SplitList(dir))
		if depth == 0 {
			// On Windows, SplitList might not work as expected, count separators manually
			for _, char := range dir {
				if char == '/' || char == filepath.Separator {
					depth++
				}
			}
		}
	}

	if depth > 0 {
		cleaned = AdjustPathsForDepth(cleaned, depth)
	}

	// Ensure output directory exists
	outputPath := filepath.Join(g.outputDir, page.OutputPath)
	destDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create page directory %s: %w", destDir, err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(cleaned), 0644); err != nil {
		return fmt.Errorf("failed to write page file: %w", err)
	}

	return nil
}

// GenerateMainSitePreview generates a preview for the main index page
func (g *MainSiteGenerator) GenerateMainSitePreview(site *Site, mainTemplateContent string, headerContent, footerContent string, previewDir string) error {
	// Ensure preview directory exists
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	// Create a template set with all components, header, and footer
	t := template.New("Main")

	_, err := t.Parse(mainTemplateContent)
	if err != nil {
		return &TemplateParseError{
			Name:   "Main",
			Source: mainTemplateContent,
			Err:    err,
		}
	}

	for name, comp := range site.Components {
		_, err := t.New(name).Parse(comp.Template)
		if err != nil {
			return &TemplateParseError{
				Name:   name,
				Source: comp.Template,
				Err:    err,
			}
		}
	}

	if headerContent != "" {
		if _, err := t.New("header.html").Parse(headerContent); err != nil {
			return &TemplateParseError{Name: "header.html", Source: headerContent, Err: err}
		}
	}

	if footerContent != "" {
		if _, err := t.New("footer.html").Parse(footerContent); err != nil {
			return &TemplateParseError{Name: "footer.html", Source: footerContent, Err: err}
		}
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, site.Data.GetAll()); err != nil {
		return &TemplateExecuteError{Name: "Main", Err: err}
	}

	cleaned := utils.CleanupWhitespace(buf.String())
	cleaned = AdjustPathsForPreview(cleaned)

	outputPath := filepath.Join(previewDir, "index.html")
	if err := os.WriteFile(outputPath, []byte(cleaned), 0644); err != nil {
		return fmt.Errorf("failed to write main site preview file: %w", err)
	}

	return nil
}

// GeneratePagePreviews generates preview pages for all pages in the preview directory
func (g *MainSiteGenerator) GeneratePagePreviews(site *Site, headerContent, footerContent string, previewDir string) error {
	// Ensure preview directory exists
	if err := os.MkdirAll(previewDir, 0755); err != nil {
		return fmt.Errorf("failed to create preview directory: %w", err)
	}

	for _, page := range site.Pages {
		if err := g.generatePagePreview(page, site, headerContent, footerContent, previewDir); err != nil {
			return fmt.Errorf("failed to generate preview for page %s: %w", page.OutputPath, err)
		}
	}
	return nil
}

// generatePagePreview generates a single page preview
func (g *MainSiteGenerator) generatePagePreview(page *Page, site *Site, headerContent, footerContent string, previewDir string) error {
	// Create a template set with all components, header, and footer
	t := template.New("Page")

	// Parse the page content
	_, err := t.Parse(page.Content)
	if err != nil {
		return &TemplateParseError{
			Name:   page.OutputPath,
			Source: page.Content,
			Err:    err,
		}
	}

	// Add components
	for name, comp := range site.Components {
		_, err := t.New(name).Parse(comp.Template)
		if err != nil {
			return &TemplateParseError{
				Name:   name,
				Source: comp.Template,
				Err:    err,
			}
		}
	}

	// Add header and footer if they exist
	if headerContent != "" {
		_, err := t.New("header.html").Parse(headerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "header.html",
				Source: headerContent,
				Err:    err,
			}
		}
	}

	if footerContent != "" {
		_, err := t.New("footer.html").Parse(footerContent)
		if err != nil {
			return &TemplateParseError{
				Name:   "footer.html",
				Source: footerContent,
				Err:    err,
			}
		}
	}

	// Execute the page template with all data
	var buf bytes.Buffer
	if err := t.Execute(&buf, site.Data.GetAll()); err != nil {
		return &TemplateExecuteError{
			Name: page.OutputPath,
			Err:  err,
		}
	}

	// Clean up excessive whitespace
	cleaned := utils.CleanupWhitespace(buf.String())

	// Adjust paths for preview directory (same as component previews)
	cleaned = AdjustPathsForPreview(cleaned)

	// Use the base filename for the preview (e.g., "google.html" not "subdir/index.html")
	previewName := filepath.Base(page.OutputPath)
	// For subdirectory pages, use the directory name instead
	if previewName == "index.html" {
		dir := filepath.Dir(page.OutputPath)
		previewName = filepath.Base(dir) + ".html"
	}

	outputPath := filepath.Join(previewDir, previewName)
	if err := os.WriteFile(outputPath, []byte(cleaned), 0644); err != nil {
		return fmt.Errorf("failed to write page preview file: %w", err)
	}

	return nil
}

// CopyAssets copies static assets to the output directory
func (g *MainSiteGenerator) CopyAssets(assets []Asset) error {
	for _, asset := range assets {
		// Ensure destination directory exists
		destDir := filepath.Dir(filepath.Join(g.outputDir, asset.OutputPath))
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return fmt.Errorf("failed to create asset directory %s: %w", destDir, err)
		}

		// Copy the file
		content, err := os.ReadFile(asset.SourcePath)
		if err != nil {
			return fmt.Errorf("failed to read asset %s: %w", asset.SourcePath, err)
		}

		destPath := filepath.Join(g.outputDir, asset.OutputPath)
		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write asset %s: %w", destPath, err)
		}
	}

	return nil
}

// CopyStylesheet copies all CSS files to the output directory
func (g *MainSiteGenerator) CopyStylesheet(rootPath string) error {
	// Find all CSS files in root directory
	cssPattern := filepath.Join(rootPath, "*.css")
	cssFiles, err := filepath.Glob(cssPattern)
	if err != nil {
		return fmt.Errorf("failed to find CSS files: %w", err)
	}

	// If no CSS files found, that's okay (they're optional)
	if len(cssFiles) == 0 {
		return nil
	}

	// Copy each CSS file
	for _, cssFile := range cssFiles {
		content, err := os.ReadFile(cssFile)
		if err != nil {
			return fmt.Errorf("failed to read CSS file %s: %w", cssFile, err)
		}

		// Get just the filename
		filename := filepath.Base(cssFile)
		destPath := filepath.Join(g.outputDir, filename)

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write CSS file %s: %w", filename, err)
		}
	}

	return nil
}
