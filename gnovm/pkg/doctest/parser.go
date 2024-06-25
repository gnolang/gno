package doctest

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// codeBlock represents a block of code extracted from the input text.
type codeBlock struct {
	content string // The content of the code block.
	start   int    // The start byte position of the code block in the input text.
	end     int    // The end byte position of the code block in the input text.
	lang    string // The language type of the code block.
	index   int    // The index of the code block in the sequence of extracted blocks.
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

	return codeBlock{
		content: content,
		start:   start,
		end:     end,
		lang:    language,
		index:   index,
	}
}
