package utils

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// ExtractBodyContent reads an HTML file and returns the content of its <body> tag as a string.
// It returns an error if the file cannot be read or if no body tag is found.
func ExtractBodyContent(filePath string) (string, error) {
	// Read the HTML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the HTML
	doc, err := html.Parse(strings.NewReader(string(data)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the body tag
	var bodyNode *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			bodyNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
			if bodyNode != nil {
				return
			}
		}
	}
	findBody(doc)

	if bodyNode == nil {
		return "", fmt.Errorf("no body tag found in HTML")
	}

	// Extract the inner content of the body tag
	var buf strings.Builder
	var extractContent func(*html.Node)
	extractContent = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
		} else if n.Type == html.ElementNode {
			// Write opening tag
			buf.WriteString("<")
			buf.WriteString(n.Data)
			for _, attr := range n.Attr {
				buf.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
			}
			buf.WriteString(">")

			// Process children
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				extractContent(c)
			}

			// Write closing tag
			buf.WriteString("</")
			buf.WriteString(n.Data)
			buf.WriteString(">")
		}
	}

	// Extract content from body's children (not including the body tag itself)
	for c := bodyNode.FirstChild; c != nil; c = c.NextSibling {
		extractContent(c)
	}

	return buf.String(), nil
}

// ExtractTemplatesAndBody reads an HTML file and returns the content of <preview> tags
// from the <head> and the content of the <body> tag as separate strings.
// It returns an error if the file cannot be read or if no body tag is found.
// Uses simple string extraction to preserve Go template syntax.
func ExtractTemplatesAndBody(filePath string) (string, string, error) {
	// Read the HTML file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)

	// Extract preview tag using simple string parsing
	dataPath := ""
	previewStart := strings.Index(content, "<preview>")
	if previewStart != -1 {
		previewEnd := strings.Index(content[previewStart:], "</preview>")
		if previewEnd != -1 {
			dataPath = content[previewStart+9 : previewStart+previewEnd]
			dataPath = strings.TrimSpace(dataPath)
			if dataPath != "" && !strings.HasPrefix(dataPath, ".") {
				dataPath = "." + dataPath
			}
		}
	}

	// Extract body content using string parsing to preserve Go template syntax
	// HTML parser corrupts template expressions like {{ . }}
	bodyStart := strings.Index(content, "<body>")
	bodyEnd := strings.Index(content, "</body>")

	if bodyStart == -1 || bodyEnd == -1 {
		return "", "", fmt.Errorf("no body tag found in HTML")
	}

	// Extract content between <body> and </body>
	bodyContent := content[bodyStart+6 : bodyEnd]

	return dataPath, bodyContent, nil
}

// CleanupWhitespace removes excessive newlines from HTML content
// It removes all blank lines that appear between tags
func CleanupWhitespace(content string) string {
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip completely empty lines
		if trimmed == "" {
			continue
		}

		result = append(result, line)
	}

	return strings.Join(result, "\n")
}
