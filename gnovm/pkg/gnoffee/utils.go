package gnoffee

import (
	"strings"
)

// normalizeGoCode normalizes a multi-line Go code string by
// trimming the common leading white spaces from each line while preserving indentation.
func normalizeGoCode(code string) string {
	code = strings.ReplaceAll(code, "\t", "        ")

	lines := strings.Split(code, "\n")

	const defaultMax = 1337 // Initialize max with an arbitrary value

	// Determine the minimum leading whitespace across all lines
	var minLeadingSpaces = defaultMax
	for _, line := range lines {
		// skip empty lines
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		// println(len(line), len(strings.TrimLeft(line, " ")), "AAA", strings.TrimLeft(line, " "), "BBB")
		if leadingSpaces < minLeadingSpaces {
			minLeadingSpaces = leadingSpaces
		}
	}
	// println(minLeadingSpaces)
	// println()

	if minLeadingSpaces == defaultMax {
		return code
	}

	// Trim the determined number of leading whitespaces from all lines
	var normalizedLines []string
	for _, line := range lines {
		if len(line) > minLeadingSpaces {
			normalizedLines = append(normalizedLines, line[minLeadingSpaces:])
		} else {
			normalizedLines = append(normalizedLines, strings.TrimSpace(line))
		}
	}

	normalizedCode := strings.Join(normalizedLines, "\n")
	normalizedCode = strings.ReplaceAll(normalizedCode, "        ", "\t")
	normalizedCode = strings.TrimSpace(normalizedCode)
	return normalizedCode
}
