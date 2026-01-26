// Package site provides high-level site orchestration.
// It coordinates loading, parsing, and generation of static sites.
package site

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"

	"genny/pkg/generator"
	"genny/pkg/loader"
	"genny/pkg/parser"

	"github.com/toolvox/utilgo/pkg/errs"
)

// Site encapsulates all site operations
type Site struct {
	rootPath    string
	site        *generator.Site
	loader      loader.Loader
	parser      *parser.ComponentParser
	tagReplacer *parser.TagReplacer

	// Cached templates
	wrapperTemplate     *template.Template
	mainTemplateContent string
	headerContent       string
	footerContent       string

	// Original content before tag replacement (for usage tracking)
	originalPageContent        map[string]string
	originalComponentTemplates map[string]string
	originalMainContent        string

	verbose bool
}

// NewSite creates a new Site
func NewSite(rootPath string, verbose bool) *Site {
	return &Site{
		rootPath:    rootPath,
		loader:      loader.NewFileSystemLoader(),
		parser:      parser.NewComponentParser(verbose),
		tagReplacer: parser.NewTagReplacer(),
		verbose:     verbose,
	}
}

// Load loads all site data from the file system
func (s *Site) Load() error {
	var siteRootPath string = s.rootPath
	if siteRootPath == "." {
		siteRootPath = errs.Must(os.Getwd())
	}
	log.Printf("Loading site from: %s", siteRootPath)

	// Load assets
	assets, err := s.loader.LoadAssets(s.rootPath)
	if err != nil {
		return fmt.Errorf("failed to load assets: %w", err)
	}
	log.Printf("Loaded %d assets", len(assets))

	// Load data
	data, err := s.loader.LoadData(s.rootPath)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}
	log.Printf("Loaded data")
	if s.verbose {
		log.Printf("data: %+v", data)
	}

	// Load components
	components, err := s.loader.LoadComponents(s.rootPath)
	if err != nil {
		return fmt.Errorf("failed to load components: %w", err)
	}
	log.Printf("Loaded %d components", len(components))

	// Parse components
	if err := s.parser.ParseComponents(components); err != nil {
		return fmt.Errorf("failed to parse components: %w", err)
	}

	// Store original component templates before tag replacement
	s.originalComponentTemplates = make(map[string]string)
	for name, comp := range components {
		s.originalComponentTemplates[name] = comp.Template
	}

	// Replace component tags
	s.tagReplacer.ReplaceComponentTagsInAllComponents(components)

	// Load pages
	pages, err := s.loader.LoadPages(s.rootPath)
	if err != nil {
		return fmt.Errorf("failed to load pages: %w", err)
	}
	log.Printf("Loaded %d pages", len(pages))

	// Load templates
	templates, err := s.loader.LoadTemplates(s.rootPath)
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Store original page content before wrapping and tag replacement
	s.originalPageContent = make(map[string]string)
	for _, page := range pages {
		s.originalPageContent[page.SourcePath] = page.Content
	}

	// Process pages - wrap with header/footer and replace component tags in page content
	for _, page := range pages {
		// Wrap page with header and footer templates
		wrapped, err := s.parser.WrapPageWithHeaderFooter(page.Content)
		if err != nil {
			return fmt.Errorf("failed to wrap page %s with header/footer: %w", page.SourcePath, err)
		}
		page.Content = wrapped

		// Replace component tags
		page.Content = s.tagReplacer.ReplaceComponentTags(page.Content, components)
	}

	// Create the Site struct
	s.site = &generator.Site{
		RootPath:   s.rootPath,
		Assets:     assets,
		Data:       generator.NewSimpleDataContext(data),
		Components: components,
		Pages:      pages,
		Templates:  make(map[string]*template.Template),
	}

	// Parse index.html to create wrapper and main templates
	indexHTML, exists := templates["index.html"]
	if !exists {
		return fmt.Errorf("index.html not found")
	}

	// Create wrapper template
	wrapperContent, err := s.parser.ExtractWrapper(indexHTML)
	if err != nil {
		return fmt.Errorf("failed to extract wrapper: %w", err)
	}

	s.wrapperTemplate, err = template.New("Wrapper").Parse(wrapperContent)
	if err != nil {
		return &generator.TemplateParseError{
			Name:   "Wrapper",
			Source: wrapperContent,
			Err:    err,
		}
	}

	// Create main template content
	mainContent, err := s.parser.ExtractMain(indexHTML)
	if err != nil {
		return fmt.Errorf("failed to extract main: %w", err)
	}

	// Store original main content
	s.originalMainContent = mainContent

	// Replace component tags in main template
	s.mainTemplateContent = s.tagReplacer.ReplaceComponentTags(mainContent, components)

	// Store header and footer content and replace component tags
	s.headerContent = s.tagReplacer.ReplaceComponentTags(templates["header.html"], components)
	s.footerContent = s.tagReplacer.ReplaceComponentTags(templates["footer.html"], components)

	log.Println("Site loaded successfully")
	return nil
}

// Generate generates the entire site
func (s *Site) Generate() error {
	if s.site == nil {
		return fmt.Errorf("site not loaded - call Load() first")
	}

	log.Println("Generating site...")

	// Track component usage
	usedComponents := s.findUsedComponents()

	// Generate component previews
	previewDir := "./www/preview"
	componentGen := generator.NewComponentGenerator(previewDir, s.verbose)
	if err := componentGen.GenerateComponentPreviews(s.site, s.wrapperTemplate); err != nil {
		return fmt.Errorf("failed to generate component previews: %w", err)
	}
	log.Printf("Generated %d component previews", len(s.site.Components))

	// Generate main site
	mainGen := generator.NewMainSiteGenerator("./www")
	if err := mainGen.GenerateMainSite(s.site, s.mainTemplateContent, s.headerContent, s.footerContent); err != nil {
		return fmt.Errorf("failed to generate main site: %w", err)
	}
	log.Println("Generated main site")

	// Generate all pages
	if err := mainGen.GeneratePages(s.site, s.headerContent, s.footerContent); err != nil {
		return fmt.Errorf("failed to generate pages: %w", err)
	}
	log.Printf("Generated %d pages", len(s.site.Pages))

	// Copy assets
	if err := mainGen.CopyAssets(s.site.Assets); err != nil {
		return fmt.Errorf("failed to copy assets: %w", err)
	}
	log.Printf("Copied %d assets", len(s.site.Assets))

	// Copy stylesheet
	if err := mainGen.CopyStylesheet(s.rootPath); err != nil {
		return fmt.Errorf("failed to copy stylesheet: %w", err)
	}
	log.Println("Copied stylesheet")

	// Report unused components
	s.reportUnusedComponents(usedComponents)

	log.Println("Site generation complete!")
	return nil
}

// GetSite returns the underlying Site struct
func (s *Site) GetSite() *generator.Site {
	return s.site
}

// findUsedComponents recursively finds all components used in pages and other components
func (s *Site) findUsedComponents() map[string]bool {
	used := make(map[string]bool)

	// Track components used in main template (use original)
	for name := range s.site.Components {
		if s.isComponentUsedInContent(name, s.originalMainContent) {
			used[name] = true
		}
	}

	// Track components used in header and footer
	for name := range s.site.Components {
		if s.isComponentUsedInContent(name, s.headerContent) ||
			s.isComponentUsedInContent(name, s.footerContent) {
			used[name] = true
		}
	}

	// Track components used in pages (use original content)
	for _, page := range s.site.Pages {
		originalContent := s.originalPageContent[page.SourcePath]
		for name := range s.site.Components {
			if s.isComponentUsedInContent(name, originalContent) {
				used[name] = true
			}
		}
	}

	// Recursively add components that are dependencies of used components
	// Use original component templates to find dependencies
	changed := true
	for changed {
		changed = false
		for name := range used {
			originalTemplate := s.originalComponentTemplates[name]
			for depName := range s.site.Components {
				if !used[depName] && s.isComponentUsedInContent(depName, originalTemplate) {
					used[depName] = true
					changed = true
				}
			}
		}
	}

	return used
}

// isComponentUsedInContent checks if a component tag appears in content
func (s *Site) isComponentUsedInContent(componentName, content string) bool {
	openTag := fmt.Sprintf("<%s>", componentName)
	return strings.Contains(content, openTag)
}

// reportUnusedComponents logs unused components
func (s *Site) reportUnusedComponents(used map[string]bool) {
	var unused []string
	for name, comp := range s.site.Components {
		if !used[name] {
			unused = append(unused, comp.FilePath)
		}
	}

	if len(unused) > 0 {
		log.Println()
		log.Println("⚠ Unused components detected:")
		for _, path := range unused {
			log.Printf("  - %s", path)
		}
		log.Println()
	}
}
