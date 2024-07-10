package doctest

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// codeBlock represents a block of code extracted from the input text.
type codeBlock struct {
	content       string // The content of the code block.
	start         int    // The start byte position of the code block in the input text.
	end           int    // The end byte position of the code block in the input text.
	lang          string // The language type of the code block.
	index         int    // The index of the code block in the sequence of extracted blocks.
	expectedOutput string // The expected output of the code block.
	expectedError  string // The expected error of the code block.
}

// GetCodeBlocks extracts all code blocks from the provided markdown text.
func GetCodeBlocks(body string) []codeBlock {
	md := goldmark.New()
	reader := text.NewReader([]byte(body))
	doc := md.Parser().Parse(reader)

	var codeBlocks []codeBlock
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if cb, ok := n.(*ast.FencedCodeBlock); ok {
				codeBlock := createCodeBlock(cb, body, len(codeBlocks))
				codeBlocks = append(codeBlocks, codeBlock)
			}
		}
		return ast.WalkContinue, nil
	})

	return codeBlocks
}

// createCodeBlock creates a CodeBlock from a goldmark FencedCodeBlock node.
func createCodeBlock(node *ast.FencedCodeBlock, body string, index int) codeBlock {
	var buf bytes.Buffer
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		buf.Write(line.Value([]byte(body)))
	}

	content := buf.String()
	language := string(node.Language([]byte(body)))
	if language == "" {
		language = "plain"
	}
	start := node.Lines().At(0).Start
	end := node.Lines().At(node.Lines().Len() - 1).Stop

	expectedOutput, expectedError, err := parseExpectedResults(content)
	if err != nil {
		panic(err)
	}

	return codeBlock{
		content:        content,
		start:          start,
		end:            end,
		lang:           language,
		index:          index,
		expectedOutput: expectedOutput,
		expectedError:  expectedError,
	}
}

func parseExpectedResults(content string) (string, string, error) {
    outputRegex := regexp.MustCompile(`(?m)^// Output:$([\s\S]*?)(?:^(?://\s*$|// Error:|$))`)
    errorRegex := regexp.MustCompile(`(?m)^// Error:$([\s\S]*?)(?:^(?://\s*$|// Output:|$))`)

    var outputs, errors []string

    cleanSection := func(section string) string {
        lines := strings.Split(section, "\n")
        var cleanedLines []string
        for _, line := range lines {
            trimmedLine := strings.TrimPrefix(line, "//")
            if len(trimmedLine) > 0 && trimmedLine[0] == ' ' {
                trimmedLine = trimmedLine[1:]
            }
            if trimmedLine != "" {
                cleanedLines = append(cleanedLines, trimmedLine)
            }
        }
        return strings.Join(cleanedLines, "\n")
    }

    outputMatches := outputRegex.FindAllStringSubmatch(content, -1)
    for _, match := range outputMatches {
        if len(match) > 1 {
            cleaned := cleanSection(match[1])
            if cleaned != "" {
                outputs = append(outputs, cleaned)
            }
        }
    }

    errorMatches := errorRegex.FindAllStringSubmatch(content, -1)
    for _, match := range errorMatches {
        if len(match) > 1 {
            cleaned := cleanSection(match[1])
            if cleaned != "" {
                errors = append(errors, cleaned)
            }
        }
    }

    expectedOutput := strings.Join(outputs, "\n")
    expectedError := strings.Join(errors, "\n")

    return expectedOutput, expectedError, nil
}