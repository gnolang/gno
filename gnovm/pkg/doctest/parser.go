package doctest

import (
	"bytes"
	"fmt"
	gast "go/ast"
	"go/parser"
	"go/token"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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
}

// GetCodeBlocks parses the provided markdown text to extract all embedded code blocks.
// It returns a slice of codeBlock structs, each representing a distinct block of code found in the markdown.
func GetCodeBlocks(body string) []codeBlock {
	md := goldmark.New()
	reader := text.NewReader([]byte(body))
	doc := md.Parser().Parse(reader)

	var codeBlocks []codeBlock
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if cb, ok := n.(*ast.FencedCodeBlock); ok {
				codeBlock := createCodeBlock(cb, body, len(codeBlocks))
				codeBlock.name = generateCodeBlockName(codeBlock.content)
				codeBlocks = append(codeBlocks, codeBlock)
			}
		}
		return ast.WalkContinue, nil
	})

	return codeBlocks
}

// createCodeBlock creates a CodeBlock from a code block node.
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

// parseExpectedResults scans the code block content for expecting outputs and errors,
// which are typically indicated by special comments in the code.
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

// generateCodeBlockName derives a name for the code block based either on special annotations within the code
// or by analyzing the code structure, such as function name or variable declaration.
func generateCodeBlockName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "// @test:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "// @test:"))
		}
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		return generateFallbackName(content)
	}

	var mainFunc *gast.FuncDecl
	for _, decl := range f.Decls {
		if fn, ok := decl.(*gast.FuncDecl); ok {
			if fn.Name.Name == "main" {
				mainFunc = fn // save the main and keep looking for a better name
			} else {
				return generateFunctionName(fn)
			}
		}
	}

	// analyze main function if it only exists
	if mainFunc != nil {
		return analyzeMainFunction(mainFunc)
	}

	// find the first top-level declaration
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *gast.GenDecl:
			if len(d.Specs) > 0 {
				switch s := d.Specs[0].(type) {
				case *gast.ValueSpec:
					if len(s.Names) > 0 {
						return s.Names[0].Name
					}
				case *gast.TypeSpec:
					return s.Name.Name
				}
			}
		}
	}

	return generateFallbackName(content)
}

// generateFunctionName creates a descriptive name for a function declaration,
// including the function name and its parameters.
func generateFunctionName(fn *gast.FuncDecl) string {
	params := make([]string, 0)
	if fn.Type.Params != nil {
		for _, param := range fn.Type.Params.List {
			paramType := ""
			if ident, ok := param.Type.(*gast.Ident); ok {
				paramType = ident.Name
			}
			for _, name := range param.Names {
				params = append(params, fmt.Sprintf("%s %s", name.Name, paramType))
			}
		}
	}
	return fmt.Sprintf("%s(%s)", fn.Name.Name, strings.Join(params, ", "))
}

// analyzeMainFunction examines the main function declaration to extract a meaningful name,
// typically based on the first significant call expression within the function body.
func analyzeMainFunction(fn *gast.FuncDecl) string {
	if fn.Body == nil {
		return "main()"
	}
	for _, stmt := range fn.Body.List {
		if exprStmt, ok := stmt.(*gast.ExprStmt); ok {
			if callExpr, ok := exprStmt.X.(*gast.CallExpr); ok {
				return generateCallExprName(callExpr)
			}
		}
	}
	return "main()"
}

// generateCallExprName constructs a name for call expression by extracting the function name
// and formatting the arguments into a readable string.
func generateCallExprName(callExpr *gast.CallExpr) string {
	funcName := ""
	if ident, ok := callExpr.Fun.(*gast.Ident); ok {
		funcName = ident.Name
	} else if selectorExpr, ok := callExpr.Fun.(*gast.SelectorExpr); ok {
		if ident, ok := selectorExpr.X.(*gast.Ident); ok {
			funcName = fmt.Sprintf("%s.%s", ident.Name, selectorExpr.Sel.Name)
		}
	}

	args := make([]string, 0)
	for _, arg := range callExpr.Args {
		if basicLit, ok := arg.(*gast.BasicLit); ok {
			args = append(args, basicLit.Value)
		} else if ident, ok := arg.(*gast.Ident); ok {
			args = append(args, ident.Name)
		}
	}

	return fmt.Sprintf("%s(%s)", funcName, strings.Join(args, ", "))
}

// generateFallbackName generates a default name for a code block when no other name could be determined.
// It uses the first significant line of the code that is not a comment or package declaration.
func generateFallbackName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "//") && trimmed != "package main" {
			if len(trimmed) > 20 {
				return trimmed[:20] + "..."
			}
			return trimmed
		}
	}
	return "unnamed_block"
}
