package doctest

import (
	"context"
	"fmt"
	"os"

	sitter "github.com/smacker/go-tree-sitter"
	markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
)

// CodeBlock represents a block of code extracted from the input text.
type CodeBlock struct {
	Content string // The content of the code block.
	Start   uint32 // The start byte position of the code block in the input text.
	End     uint32 // The end byte position of the code block in the input text.
	T       string // The language type of the code block.
	Index   int    // The index of the code block in the sequence of extracted blocks.
}

// ReadMarkdownFile reads a markdown file and returns its content
func ReadMarkdownFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// getCodeBlocks extracts all code blocks from the provided markdown text.
func getCodeBlocks(body string) []CodeBlock {
	parser := createParser()
	tree, err := parseMarkdown(parser, body)
	if err != nil {
		fmt.Println("Error parsing:", err)
		return nil
	}

	return extractCodeBlocks(tree.RootNode(), body)
}

// createParser creates and returns a new tree-sitter parser configured for Markdown.
func createParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(markdown.GetLanguage())
	return parser
}

// parseMarkdown parses the input markdown text and returns the parse tree.
func parseMarkdown(parser *sitter.Parser, body string) (*sitter.Tree, error) {
	ctx := context.Background()
	return parser.ParseCtx(ctx, nil, []byte(body))
}

// extractCodeBlocks traverses the parse tree and extracts code blocks.
func extractCodeBlocks(rootNode *sitter.Node, body string) []CodeBlock {
	codeBlocks := []CodeBlock{}
	var index int

	var extract func(node *sitter.Node)
	extract = func(node *sitter.Node) {
		if node == nil {
			return
		}

		if node.Type() == "code_fence_content" {
			codeBlock := createCodeBlock(node, body, index)
			codeBlocks = append(codeBlocks, codeBlock)
			index++
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			extract(child)
		}
	}

	extract(rootNode)
	return codeBlocks
}

// createCodeBlock creates a CodeBlock from a code fence content node.
func createCodeBlock(node *sitter.Node, body string, index int) CodeBlock {
	startByte := node.StartByte()
	endByte := node.EndByte()
	content := body[startByte:endByte]

	language := detectLanguage(node, body)
	content = removeTrailingBackticks(content)

	return CodeBlock{
		Content: content,
		Start:   startByte,
		End:     endByte,
		T:       language,
		Index:   index,
	}
}

// detectLanguage detects the language of a code block from its parent node.
func detectLanguage(node *sitter.Node, body string) string {
	codeFenceNode := node.Parent()
	if codeFenceNode != nil && codeFenceNode.ChildCount() > 1 {
		langNode := codeFenceNode.Child(1)
		if langNode != nil && langNode.Type() == "info_string" {
			return langNode.Content([]byte(body))
		}
	}
	return "plain"
}

// removeTrailingBackticks removes trailing backticks from the code content.
func removeTrailingBackticks(content string) string {
	if len(content) >= 3 && content[len(content)-3:] == "```" {
		return content[:len(content)-3]
	}
	return content
}

// writeCodeBlockToFile writes a extracted code block to a temp file.
// This generated file will be executed by gnovm.
func writeCodeBlockToFile(c CodeBlock) error {
	if c.T == "go" {
		c.T = "gno"
	}

	fileName := fmt.Sprintf("%d.%s", c.Index, c.T)
	file, err := os.Create(fileName) // TODO: use temp file
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(c.Content)
	if err != nil {
		return err
	}

	return nil
}
