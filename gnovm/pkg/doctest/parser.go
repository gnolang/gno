package doctest

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	mast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// Directive names recognized in code-block comments. The grammar
// follows filetest (`gnovm/pkg/test/filetest.go`):
//
//   - PascalCase markers (Output, Error) start a multiline section
//     captured from following `// xxx` comment lines until the next
//     section marker, a bare `//`, or any non-comment line.
//   - ALLCAPS keys (NAME, IGNORE, SHOULD_PANIC) are single-line:
//     `KEY:` for a flag, `KEY: value` for a value.
const (
	directiveOutput      = "Output"
	directiveError       = "Error"
	directiveName        = "NAME"
	directiveIgnore      = "IGNORE"
	directiveShouldPanic = "SHOULD_PANIC"
)

type codeBlock struct {
	content        string
	lang           string
	index          int
	expectedOutput string
	expectedError  string
	name           string
	options        ExecutionOptions
}

type ExecutionOptions struct {
	Ignore       bool
	ShouldPanic  bool
	PanicMessage string
}

// GetCodeBlocks parses markdown text and returns every fenced code
// block, with directives and options resolved.
func GetCodeBlocks(body string) ([]codeBlock, error) {
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader([]byte(body)))

	var blocks []codeBlock
	if err := mast.Walk(doc, func(n mast.Node, entering bool) (mast.WalkStatus, error) {
		if !entering {
			return mast.WalkContinue, nil
		}
		cb, ok := n.(*mast.FencedCodeBlock)
		if !ok {
			return mast.WalkContinue, nil
		}
		blocks = append(blocks, buildCodeBlock(cb, body, len(blocks)))
		return mast.WalkContinue, nil
	}); err != nil {
		return nil, err
	}
	return blocks, nil
}

func buildCodeBlock(node *mast.FencedCodeBlock, body string, index int) codeBlock {
	var buf bytes.Buffer
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.WriteString(body[line.Start:line.Stop])
	}
	content := buf.String()

	language := string(node.Language([]byte(body)))
	if language == "" {
		language = "plain"
	}

	meta := parseBlockMetadata(content)
	if meta.name == "" {
		meta.name = fmt.Sprintf("block_%d", index)
	}
	return codeBlock{
		content:        content,
		lang:           language,
		index:          index,
		expectedOutput: meta.output,
		expectedError:  meta.errOutput,
		name:           meta.name,
		options:        meta.options,
	}
}

type blockMetadata struct {
	name      string
	output    string
	errOutput string
	options   ExecutionOptions
}

type sectionState int

const (
	sectionNone sectionState = iota
	sectionOutput
	sectionError
)

// parseBlockMetadata walks the block once, dispatching directive
// lines and capturing Output/Error sections.
func parseBlockMetadata(content string) blockMetadata {
	var (
		meta    blockMetadata
		outputs []string
		errors  []string
	)
	state := sectionNone

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		trimmed := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(trimmed, "//") {
			state = sectionNone
			continue
		}
		body := strings.TrimPrefix(trimmed, "//")
		bodyTrim := strings.TrimSpace(body)
		if bodyTrim == "" {
			state = sectionNone
			continue
		}

		if name, value, ok := parseDirective(bodyTrim); ok {
			switch name {
			case directiveOutput:
				state = sectionOutput
			case directiveError:
				state = sectionError
			case directiveName:
				if value != "" {
					meta.name = value
				}
			case directiveIgnore:
				meta.options.Ignore = true
			case directiveShouldPanic:
				meta.options.ShouldPanic = true
				if value != "" {
					meta.options.PanicMessage = value
				}
			}
			continue
		}

		switch state {
		case sectionOutput:
			outputs = append(outputs, strings.TrimPrefix(body, " "))
		case sectionError:
			errors = append(errors, strings.TrimPrefix(body, " "))
		}
	}

	meta.output = strings.Join(outputs, "\n")
	meta.errOutput = strings.Join(errors, "\n")
	return meta
}

// parseDirective recognizes a comment body of the form `KEY:` or
// `KEY: value` where KEY is one of the known directives.
func parseDirective(s string) (name, value string, ok bool) {
	key, val, hasColon := strings.Cut(s, ":")
	if !hasColon || !isDirective(key) {
		return "", "", false
	}
	return key, strings.TrimSpace(val), true
}

func isDirective(k string) bool {
	switch k {
	case directiveOutput, directiveError,
		directiveName, directiveIgnore, directiveShouldPanic:
		return true
	}
	return false
}
