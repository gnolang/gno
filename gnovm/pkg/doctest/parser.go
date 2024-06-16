package doctest

import (
	"context"
	"fmt"
	"os"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	markdown "github.com/smacker/go-tree-sitter/markdown/tree-sitter-markdown"
)

// tree-sitter node types for markdown code blocks.
// https://github.com/smacker/go-tree-sitter/blob/0ac8d7d185ec65349d3d9e6a7a493b81ae05d198/markdown/tree-sitter-markdown/scanner.c#L9-L88
const (
	FENCED_CODE_BLOCK        = "fenced_code_block"
	CODE_FENCE_CONTENT       = "code_fence_content"
	CODE_FENCE_END           = "code_fence_end"
	CODE_FENCE_END_BACKTICKS = "code_fence_end_backticks"
	INFO_STRING              = "info_string"
)

// Code block markers.
const (
	Backticks = "```"
	Tildes    = "~~~"
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

// extractCodeBlocks traverses the parse tree and extracts code blocks using tree-sitter.
// It takes the root node of the parse tree and the complete body string as input.
func extractCodeBlocks(rootNode *sitter.Node, body string) []CodeBlock {
	codeBlocks := make([]CodeBlock, 0)

	// define a recursive function to traverse the parse tree
	var traverse func(node *sitter.Node)
	traverse = func(node *sitter.Node) {
		if node.Type() == CODE_FENCE_CONTENT {
			codeBlock := createCodeBlock(node, body, len(codeBlocks))
			codeBlocks = append(codeBlocks, codeBlock)
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			traverse(child)
		}
	}

	traverse(rootNode)
	return codeBlocks
}

// createCodeBlock creates a CodeBlock from a code fence content node.
func createCodeBlock(node *sitter.Node, body string, index int) CodeBlock {
	startByte := node.StartByte()
	endByte := node.EndByte()
	content := body[startByte:endByte]

	language := detectLanguage(node, body)
	startByte, endByte, content = adjustContentBoundaries(node, startByte, endByte, content, body)

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
		if langNode != nil && langNode.Type() == INFO_STRING {
			return langNode.Content([]byte(body))
		}
	}

	// default to plain text if no language is specified
	return "plain"
}

// removeTrailingBackticks removes trailing backticks from the code content.
func removeTrailingBackticks(content string) string {
	// https://www.markdownguide.org/extended-syntax/#fenced-code-blocks
	// a code block can have a closing fence with three or more backticks or tildes.
	content = strings.TrimRight(content, "`~")
	if len(content) >= 3 {
		blockSuffix := content[len(content)-3:]
		switch blockSuffix {
		case Backticks, Tildes:
			return content[:len(content)-3]
		default:
			return content
		}
	}
	return content
}

// adjustContentBoundaries adjusts the content boundaries of a code block node.
// The function checks the parent node type and adjusts the end byte position if it is a fenced code block.
func adjustContentBoundaries(node *sitter.Node, startByte, endByte uint32, content, body string) (uint32, uint32, string) {
	parentNode := node.Parent()
	if parentNode == nil {
		return startByte, endByte, removeTrailingBackticks(content)
	}

	// adjust the end byte based on the parent node type
	if parentNode.Type() == FENCED_CODE_BLOCK {
		// find the end marker node
		endMarkerNode := findEndMarkerNode(parentNode)
		if endMarkerNode != nil {
			endByte = endMarkerNode.StartByte()
			content = body[startByte:endByte]
		}
	}

	return startByte, endByte, removeTrailingBackticks(content)
}

// findEndMarkerNode finds the end marker node of a fenced code block using tree-sitter.
// It takes the parent node of the code block as input and iterates through its child nodes.
func findEndMarkerNode(parentNode *sitter.Node) *sitter.Node {
	for i := 0; i < int(parentNode.ChildCount()); i++ {
		child := parentNode.Child(i)
		switch child.Type() {
		case CODE_FENCE_END, CODE_FENCE_END_BACKTICKS:
			return child
		default:
			continue
		}
	}

	return nil
}
