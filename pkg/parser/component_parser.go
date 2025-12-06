// Package parser provides HTML and template parsing functionality.
// It handles extracting data paths from component files, parsing Go templates,
// and converting custom component tags to template syntax.
package parser

import (
	"fmt"
	"os"

	"genny/pkg/generator"
	"genny/pkg/utils"
)

// ComponentParser handles parsing component files
type ComponentParser struct {
	verbose bool
}

// NewComponentParser creates a new ComponentParser
func NewComponentParser(verbose bool) *ComponentParser {
	return &ComponentParser{verbose: verbose}
}

// ParseComponent reads a component file and extracts its template and data path
func (p *ComponentParser) ParseComponent(comp *generator.Component) error {
	if comp.FilePath == "" {
		return fmt.Errorf("component %s has no file path", comp.Name)
	}

	// Use the existing utility to extract templates and body
	dataPath, body, err := utils.ExtractTemplatesAndBody(comp.FilePath)
	if err != nil {
		return fmt.Errorf("failed to parse component %s: %w", comp.Name, err)
	}

	if p.verbose {
		fmt.Printf("DEBUG ParseComponent: %s extracted DataPath: '%s'\n", comp.Name, dataPath)
		fmt.Printf("DEBUG %s Template length: %d chars\n", comp.Name, len(body))
	}

	comp.Template = body
	comp.DataPath = dataPath

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ParseComponents parses all components in the map
func (p *ComponentParser) ParseComponents(components map[string]*generator.Component) error {
	for name, comp := range components {
		if err := p.ParseComponent(comp); err != nil {
			return fmt.Errorf("failed to parse component %s: %w", name, err)
		}
	}
	return nil
}

// ExtractWrapper extracts the wrapper template from index.html
func (p *ComponentParser) ExtractWrapper(indexHTML string) (string, error) {
	// This is similar to GetBasicWrapper in the old code
	// Read the index.html content (passed as string)
	// Split on <body> tag and create wrapper template

	parts := splitHTMLBody(indexHTML)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid index.html structure: expected head, body, tail")
	}

	wrapper := fmt.Sprintf("%s<body>\n\t{{ . }}\n</body>%s", parts[0], parts[2])
	return wrapper, nil
}

// ExtractMain extracts the main template from index.html with header and footer
func (p *ComponentParser) ExtractMain(indexHTML string) (string, error) {
	parts := splitHTMLBody(indexHTML)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid index.html structure: expected head, body, tail")
	}

	main := fmt.Sprintf(`%s<body>
	{{ template "header.html" . }}
	%s
	{{ template "footer.html" . }}
</body>%s`, parts[0], parts[1], parts[2])

	return main, nil
}

// WrapPageWithHeaderFooter wraps page HTML content with header and footer templates
func (p *ComponentParser) WrapPageWithHeaderFooter(pageHTML string) (string, error) {
	parts := splitHTMLBody(pageHTML)
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid page HTML structure: expected head, body, tail")
	}

	wrapped := fmt.Sprintf(`%s<body>
	{{ template "header.html" . }}
	%s
	{{ template "footer.html" . }}
</body>%s`, parts[0], parts[1], parts[2])

	return wrapped, nil
}

// splitHTMLBody splits HTML into [head, body content, tail]
func splitHTMLBody(html string) []string {
	// Find <body> and </body> tags
	bodyStart := findTag(html, "<body>")
	bodyEnd := findTag(html, "</body>")

	if bodyStart == -1 || bodyEnd == -1 {
		return nil
	}

	head := html[:bodyStart]
	body := html[bodyStart+6 : bodyEnd] // +6 for "<body>"
	tail := html[bodyEnd+7:]            // +7 for "</body>"

	return []string{head, body, tail}
}

// findTag finds the index of a tag in HTML
func findTag(html, tag string) int {
	for i := 0; i < len(html)-len(tag); i++ {
		if html[i:i+len(tag)] == tag {
			return i
		}
	}
	return -1
}

// LoadAndParseTemplateFile loads a template file and returns its content
func (p *ComponentParser) LoadAndParseTemplateFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", &generator.FileNotFoundError{Path: path}
	}
	return string(content), nil
}
