package doctest

import (
	"regexp"
	"strings"
)

type CodeBlock struct {
	Content string
	Start   int
	End     int
	T       string
	Index   int
}

func getCodeBlocks(body string) []CodeBlock {
	var results []CodeBlock

	blocksRegex := regexp.MustCompile("```\\w*[^`]+```*")
	matches := blocksRegex.FindAllStringIndex(body, -1)

	// initialize index to 0. will increment for each code block
	index := 0

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		codeStr := body[match[0]:match[1]]
		// Remove the backticks from the code block content
		codeStr = strings.TrimPrefix(codeStr, "```")
		codeStr = strings.TrimSuffix(codeStr, "```")
		result := CodeBlock{
			Content: codeStr,
			Start:   match[0],
			End:     match[1],
			Index:   index, // set the current index
		}

		// extract the type (language) of the code block
		lines := strings.Split(codeStr, "\n")
		if len(lines) > 0 {
			line1 := lines[0]
			languageRegex := regexp.MustCompile(`^\w*`)
			languageMatch := languageRegex.FindString(line1)
			result.T = languageMatch
			// Remove the language specifier from the code block content
			result.Content = strings.TrimPrefix(result.Content, languageMatch)
			result.Content = strings.TrimSpace(result.Content)
		}
		if result.T == "" {
			result.T = "plain"
		}
		results = append(results, result)
		index++
	}

	return results
}
