package doctest

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	mast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const testPrefix = "// @test:"

var (
	outputRegex = regexp.MustCompile(`(?m)^// Output:$([\s\S]*?)(?:^(?://\s*$|// Error:|$))`)
	errorRegex  = regexp.MustCompile(`(?m)^// Error:$([\s\S]*?)(?:^(?://\s*$|// Output:|$))`)
)

// codeBlock represents a block of code extracted from the input text.
type codeBlock struct {
	content        string // The content of the code block.
	start          int    // The start byte position of the code block in the input text.
	end            int    // The end byte position of the code block in the input text.
	lang           string // The language type of the code block.
	index          int    // The index of the code block in the sequence of extracted blocks.
	expectedOutput string // The expected output of the code block.
	expectedError  string // The expected error of the code block.
	name           string // The name of the code block.
	options        ExecutionOptions
}

// GetCodeBlocks parses the provided markdown text to extract all embedded code blocks.
// It returns a slice of codeBlock structs, each representing a distinct block of code found in the markdown.
func GetCodeBlocks(body string) ([]codeBlock, error) {
	md := goldmark.New()
	reader := text.NewReader([]byte(body))
	doc := md.Parser().Parse(reader)

	var codeBlocks []codeBlock
	if err := mast.Walk(doc, func(n mast.Node, entering bool) (mast.WalkStatus, error) {
		if entering {
			if cb, ok := n.(*mast.FencedCodeBlock); ok {
				codeBlock, err := createCodeBlock(cb, body, len(codeBlocks))
				if err != nil {
					return mast.WalkStop, err
				}
				codeBlock.name = generateCodeBlockName(codeBlock.content, codeBlock.expectedOutput)
				codeBlocks = append(codeBlocks, codeBlock)
			}
		}
		return mast.WalkContinue, nil
	}); err != nil {
		return nil, err
	}

	return codeBlocks, nil
}

// createCodeBlock creates a CodeBlock from a code block node.
func createCodeBlock(node *mast.FencedCodeBlock, body string, index int) (codeBlock, error) {
	var buf bytes.Buffer
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write([]byte(body[line.Start:line.Stop]))
	}

	content := buf.String()
	language := string(node.Language([]byte(body)))
	if language == "" {
		language = "plain"
	}

	firstLine := body[lines.At(0).Start:lines.At(0).Stop]
	options := parseExecutionOptions(language, []byte(firstLine))

	start := lines.At(0).Start
	end := lines.At(node.Lines().Len() - 1).Stop

	expectedOutput, expectedError, err := parseExpectedResults(content)
	if err != nil {
		return codeBlock{}, err
	}

	return codeBlock{
		content:        content,
		start:          start,
		end:            end,
		lang:           language,
		index:          index,
		expectedOutput: expectedOutput,
		expectedError:  expectedError,
		options:        options,
	}, nil
}

// parseExpectedResults scans the code block content for expecting outputs and errors,
// which are typically indicated by special comments in the code.
func parseExpectedResults(content string) (string, string, error) {
	var outputs, errors []string

	outputMatches := outputRegex.FindAllStringSubmatch(content, -1)
	for _, match := range outputMatches {
		if len(match) > 1 {
			cleaned, err := cleanSection(match[1])
			if err != nil {
				return "", "", err
			}
			if cleaned != "" {
				outputs = append(outputs, cleaned)
			}
		}
	}

	errorMatches := errorRegex.FindAllStringSubmatch(content, -1)
	for _, match := range errorMatches {
		if len(match) > 1 {
			cleaned, err := cleanSection(match[1])
			if err != nil {
				return "", "", err
			}
			if cleaned != "" {
				errors = append(errors, cleaned)
			}
		}
	}

	expectedOutput := strings.Join(outputs, "\n")
	expectedError := strings.Join(errors, "\n")

	return expectedOutput, expectedError, nil
}

func cleanSection(section string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(section))
	var cleanedLines []string

	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "//"))
		line = strings.TrimSpace(line)
		if line != "" {
			cleanedLines = append(cleanedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to clean section: %w", err)
	}

	return strings.Join(cleanedLines, "\n"), nil
}

//////////////////// Auto-Name Generator ////////////////////

// generateCodeBlockName generates a name for a given code block based on its content.
// It first checks for a custom name specified with `// @test:` comment.
// If not found, it analyzes the code structure to create meaningful name.
// The name is constructed based on the code's prefix (Test, Print or Calc),
// imported packages, main identifier, and expected output.
func generateCodeBlockName(content string, expectedOutput string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), testPrefix) {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), testPrefix))
		}
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return generateFallbackName(content)
	}

	prefix := determinePrefix(f)
	imports := extractImports(f)
	mainIdentifier := extractMainIdentifier(f)

	name := constructName(prefix, imports, expectedOutput, mainIdentifier)

	return name
}

// determinePrefix analyzes the AST of a file and determines an appropriate prefix
// for the code block name.
// It returns "Test" for test functions, "Print" for functions containing print statements,
// "Calc" for function containing calculations, or an empty string if no specific prefix is determined.
func determinePrefix(f *ast.File) string {
	// determine the prefix by using heuristic
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Test") {
				return "Test"
			}
			if containsPrintStmt(fn) {
				return "Print"
			}
			if containsCalculation(fn) {
				return "Calc"
			}
		}
	}
	return ""
}

// containsPrintStmt checks if the given function declaration contains
// any print or println statements.
func containsPrintStmt(fn *ast.FuncDecl) bool {
	hasPrintStmt := false
	ast.Inspect(fn, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "println" || ident.Name == "print" {
					hasPrintStmt = true
					return false
				}
			}
		}
		return true
	})
	return hasPrintStmt
}

// containsCalculation checks if the given function declaration contains
// any binary or unary expressions, which are indicative of calculations.
func containsCalculation(fn *ast.FuncDecl) bool {
	var hasCalcExpr bool
	ast.Inspect(fn, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.BinaryExpr, *ast.UnaryExpr:
			hasCalcExpr = true
			return false
		}
		return true
	})
	return hasCalcExpr
}

// extractImports extracts the names of imported packages from the AST
// of a Go file. It returns a slice of strings representing the imported
// package names or the last part of the import path if no alias is used.
func extractImports(f *ast.File) []string {
	imports := make([]string, 0, len(f.Imports))
	for _, imp := range f.Imports {
		if imp.Name != nil {
			imports = append(imports, imp.Name.Name)
			continue
		}
		path := strings.Trim(imp.Path.Value, `"`)
		parts := strings.Split(path, "/")
		imports = append(imports, parts[len(parts)-1])
	}
	return imports
}

// extractMainIdentifier attempts to find the main identifier in the Go file.
// It returns the name of the first function or the first declared variable.
// If no suitable identifier is found, it returns an empty string.
func extractMainIdentifier(f *ast.File) string {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			return d.Name.Name
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					if len(vs.Names) > 0 {
						return vs.Names[0].Name
					}
				}
			}
		}
	}
	return ""
}

// constructName builds a name for the code block using the provided components.
// The resulting name is truncated if it exceeds 50 characters.
func constructName(
	prefix string,
	imports []string,
	expectedOutput string,
	mainIdentifier string,
) string {
	var parts []string
	if prefix != "" {
		parts = append(parts, prefix)
	}
	if mainIdentifier != "" {
		parts = append(parts, mainIdentifier)
	}
	if expectedOutput != "" {
		// use first line of expected output, limit the length
		outputPart := strings.Split(expectedOutput, "\n")[0]
		if len(outputPart) > 20 {
			outputPart = outputPart[:20] + "..."
		}
		parts = append(parts, outputPart)
	}

	// Add imports last, limiting to a certain number of characters
	if len(imports) > 0 {
		importString := strings.Join(imports, "_")
		if len(importString) > 30 {
			importString = importString[:30] + "..."
		}
		parts = append(parts, importString)
	}

	name := strings.Join(parts, "_")
	if len(name) > 50 {
		name = name[:50] + "..."
	}

	return name
}

// generateFallbackName generates a default name for a code block when no other name could be determined.
// It uses the first significant line of the code that is not a comment or package declaration.
func generateFallbackName(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && trimmed != "package main" {
			if len(trimmed) > 20 {
				return trimmed[:20] + "..."
			}
			return trimmed
		}
	}
	return "unnamed_block"
}

//////////////////// Execution Options ////////////////////

type ExecutionOptions struct {
	Ignore       bool
	PanicMessage string
	// TODO: add more options
}

func parseExecutionOptions(language string, firstLine []byte) ExecutionOptions {
	var options ExecutionOptions

	parts := strings.Split(language, ",")
	for _, option := range parts[1:] { // skip the first part which is the language
		switch strings.TrimSpace(option) {
		case "ignore":
			options.Ignore = true
		case "should_panic":
			// specific panic message will be parsed later
		}
	}

	// parser options from the first line of the code block
	if bytes.HasPrefix(firstLine, []byte("//")) {
		// parse execution options from the first line of the code block
		// e.g. // @should_panic="some panic message here"
		//        |-option name-||-----option value-----|
		re := regexp.MustCompile(`@(\w+)(?:="([^"]*)")?`)
		matches := re.FindAllSubmatch(firstLine, -1)
		for _, match := range matches {
			switch string(match[1]) {
			case "should_panic":
				if match[2] != nil {
					options.PanicMessage = string(match[2])
				}
			case "ignore":
				options.Ignore = true
			}
		}
	}

	return options
}
