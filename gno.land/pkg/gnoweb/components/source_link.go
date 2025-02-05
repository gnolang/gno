package components

import (
	"fmt"
	"html"
	"strings"
)

func ProcessGnoImports(source []byte) string {
	lines := strings.Split(string(source), "\n")
	inImportBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for import block
		if strings.HasPrefix(trimmed, "import (") {
			inImportBlock = true
			continue
		}

		if inImportBlock && trimmed == ")" {
			inImportBlock = false
			continue
		}

		// Process single line imports or imports within a block
		if inImportBlock || strings.HasPrefix(trimmed, "import ") {
			if path := extractGnoPath(trimmed); path != "" {
				lines[i] = wrapWithLink(line, path)
			}
		}
	}

	return strings.Join(lines, "\n")
}

func extractGnoPath(line string) string {
	line = strings.TrimSpace(line)

	// Handle quoted imports
	if strings.Contains(line, "\"gno.land/") {
		parts := strings.Split(line, "\"")
		for _, part := range parts {
			if strings.HasPrefix(part, "gno.land/") {
				return part
			}
		}
	}
	return ""
}

func wrapWithLink(line, path string) string {
	escaped := html.EscapeString(path)
	urlPath := strings.TrimPrefix(path, "gno.land/")

	// paths should link to $source
	link := fmt.Sprintf("<a href=\"/%s$source\" class=\"gno-import\">%s</a>", urlPath, escaped)
	return strings.Replace(line, path, link, 1)
}
