package generator

import (
	"regexp"
	"strings"
)

// AdjustPathsForPreview adjusts asset and stylesheet paths for the preview directory
// Preview files are in www/preview/, so they need to go up two levels to reach root
func AdjustPathsForPreview(html string) string {
	result := html

	// Adjust all relative href paths in link tags (not starting with /, http, https, or ../)
	// Match paths that start with alphanumeric but exclude URLs (no :// within first 10 chars)
	hrefPattern := regexp.MustCompile(`((?:link|a)[^>]*href=")([a-zA-Z][^":]*?)(")`)
	result = hrefPattern.ReplaceAllString(result, `${1}../../${2}${3}`)

	// Adjust all relative src paths (not starting with /, http, https, or ../)
	srcPattern := regexp.MustCompile(`(src=")([a-zA-Z][^":]*?)(")`)
	result = srcPattern.ReplaceAllString(result, `${1}../../${2}${3}`)

	return result
}

// AdjustPathsForDepth adjusts paths based on directory depth from root
func AdjustPathsForDepth(html string, depth int) string {
	if depth == 0 {
		return html
	}

	result := html
	prefix := strings.Repeat("../", depth)

	// Adjust all relative href paths in link tags (not starting with /, http, https, or ../)
	hrefPattern := regexp.MustCompile(`((?:link|a)[^>]*href=")([a-zA-Z][^":]*?)(")`)
	result = hrefPattern.ReplaceAllString(result, `${1}`+prefix+`${2}${3}`)

	// Adjust all relative src paths (not starting with /, http, https, or ../)
	srcPattern := regexp.MustCompile(`(src=")([a-zA-Z][^":]*?)(")`)
	result = srcPattern.ReplaceAllString(result, `${1}`+prefix+`${2}${3}`)

	return result
}
