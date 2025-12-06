package parser

import (
	"fmt"
	"strings"

	"genny/pkg/generator"
)

// TagReplacer handles converting custom component tags to Go template syntax
type TagReplacer struct{}

// NewTagReplacer creates a new TagReplacer
func NewTagReplacer() *TagReplacer {
	return &TagReplacer{}
}

// ReplaceComponentTags converts <component>path</component> to {{ template "component" path }}
func (r *TagReplacer) ReplaceComponentTags(templateContent string, components map[string]*generator.Component) string {
	result := templateContent

	for name := range components {
		openTag := fmt.Sprintf("<%s>", name)
		closeTag := fmt.Sprintf("</%s>", name)

		if !strings.Contains(result, openTag) {
			continue
		}

		templateOpen := fmt.Sprintf(`{{ template "%s" `, name)
		templateClose := " }}"

		result = strings.ReplaceAll(result, openTag, templateOpen)
		result = strings.ReplaceAll(result, closeTag, templateClose)
	}

	// Remove any remaining <preview> tags from the template
	result = r.RemovePreviewTags(result)

	return result
}

// RemovePreviewTags removes all <preview>...</preview> tags from the template
func (r *TagReplacer) RemovePreviewTags(templateContent string) string {
	result := templateContent

	// Remove preview tags and their content
	for {
		start := strings.Index(result, "<preview>")
		if start == -1 {
			break
		}

		end := strings.Index(result[start:], "</preview>")
		if end == -1 {
			break
		}

		// Remove the entire <preview>...</preview> block
		result = result[:start] + result[start+end+len("</preview>"):]
	}

	return result
}

// ExtractComponentDependencies finds all component references in a template
func (r *TagReplacer) ExtractComponentDependencies(templateContent string, components map[string]*generator.Component) []string {
	var dependencies []string

	for name := range components {
		openTag := fmt.Sprintf("<%s>", name)
		if strings.Contains(templateContent, openTag) {
			dependencies = append(dependencies, name)
		}
	}

	return dependencies
}

// ReplaceComponentTagsInAllComponents processes all components and replaces their tags
func (r *TagReplacer) ReplaceComponentTagsInAllComponents(components map[string]*generator.Component) {
	// First pass: extract dependencies
	for _, comp := range components {
		comp.Dependencies = r.ExtractComponentDependencies(comp.Template, components)
	}

	// Second pass: replace tags
	for _, comp := range components {
		comp.Template = r.ReplaceComponentTags(comp.Template, components)
	}
}
