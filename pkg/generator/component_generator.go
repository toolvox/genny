package generator

import (
	"bytes"
	"fmt"
	"genny/pkg/utils"
	"html/template"
	"os"
	"path/filepath"
)

// ComponentGenerator handles generating component previews
type ComponentGenerator struct {
	outputDir string
	verbose   bool
}

// NewComponentGenerator creates a new ComponentGenerator
func NewComponentGenerator(outputDir string, verbose bool) *ComponentGenerator {
	return &ComponentGenerator{outputDir: outputDir, verbose: verbose}
}

// GenerateComponentPreviews generates preview pages for all components
func (g *ComponentGenerator) GenerateComponentPreviews(site *Site, wrapperTemplate *template.Template) error {
	// Ensure output directory exists
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create a template set with all components
	t := template.New("components")
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

	// Generate preview for each component
	for name, comp := range site.Components {
		if err := g.generateComponentPreview(name, comp, t, wrapperTemplate, site.Data); err != nil {
			return fmt.Errorf("failed to generate preview for component %s: %w", name, err)
		}
	}

	return nil
}

// generateComponentPreview generates a single component preview
func (g *ComponentGenerator) generateComponentPreview(name string, comp *Component, templateSet *template.Template, wrapperTemplate *template.Template, dataContext DataContext) error {
	if g.verbose {
		fmt.Printf("DEBUG: Component %s has DataPath: '%s'\n", name, comp.DataPath)
	}

	// Get the data for this component
	data, err := dataContext.Get(comp.DataPath)
	if err != nil {
		return fmt.Errorf("failed to get data for component %s at path %s: %w", name, comp.DataPath, err)
	}

	if g.verbose {
		fmt.Printf("DEBUG: Component %s got data of type: %T\n", name, data)
	}

	// Execute the component template
	var componentBuf bytes.Buffer
	componentTmpl := templateSet.Lookup(name)
	if componentTmpl == nil {
		return fmt.Errorf("component template not found: %s", name)
	}

	if err := componentTmpl.Execute(&componentBuf, data); err != nil {
		return &TemplateExecuteError{
			Name: name,
			Err:  err,
		}
	}

	// Wrap the component in the wrapper template
	var resultBuf bytes.Buffer
	if err := wrapperTemplate.Execute(&resultBuf, template.HTML(componentBuf.String())); err != nil {
		return &TemplateExecuteError{
			Name: "Wrapper",
			Err:  err,
		}
	}

	// Adjust paths for preview directory
	result := AdjustPathsForPreview(resultBuf.String())

	// Clean up excessive whitespace
	result = utils.CleanupWhitespace(result)

	// Write to file
	filename := filepath.Join(g.outputDir, fmt.Sprintf("%s.html", name))
	if err := os.WriteFile(filename, []byte(result), 0644); err != nil {
		return fmt.Errorf("failed to write preview file: %w", err)
	}

	return nil
}
