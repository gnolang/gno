package markdown

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

// Filter provides text content filtering for Markdown rendering.
type Filter struct {
	mu                 sync.RWMutex
	patterns           []*compiledPattern
	defaultReplacement string
}

type compiledPattern struct {
	regex       *regexp.Regexp
	replacement string
}

func NewFilter(filePath string) (*Filter, error) {
	f := &Filter{
		defaultReplacement: "[filtered]",
	}
	if err := f.loadPatterns(filePath); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Filter) loadPatterns(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open pattern file: %w", err)
	}
	defer file.Close()

	var patterns []*compiledPattern
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "DEFAULT_REPLACEMENT=") {
			f.defaultReplacement = strings.TrimPrefix(line, "DEFAULT_REPLACEMENT=")
			continue
		}

		var patternStr, replacement string
		if idx := strings.Index(line, " → "); idx != -1 {
			patternStr = strings.TrimSpace(line[:idx])
			replacement = strings.TrimSpace(line[idx+len(" → "):])
		} else {
			patternStr = line
			replacement = ""
		}

		regex, err := regexp.Compile(patternStr)
		if err != nil {
			continue
		}

		patterns = append(patterns, &compiledPattern{
			regex:       regex,
			replacement: replacement,
		})
	}

	f.mu.Lock()
	f.patterns = patterns
	f.mu.Unlock()

	return scanner.Err()
}

func (f *Filter) FilterText(text string) string {
	if f == nil {
		return text
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	filtered := text
	for _, p := range f.patterns {
		replacement := p.replacement
		if replacement == "" {
			replacement = f.defaultReplacement
		}
		filtered = p.regex.ReplaceAllString(filtered, replacement)
	}
	return filtered
}

type contentFilterRenderer struct {
	filter *Filter
}

func (r *contentFilterRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindText, r.renderText)
}

func (r *contentFilterRenderer) renderText(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Text)
	segment := n.Segment
	value := string(segment.Value(source))

	filtered := r.filter.FilterText(value)

	if n.IsRaw() {
		_, _ = w.Write([]byte(filtered))
	} else {
		_, _ = w.WriteString(html.EscapeString(filtered))
	}

	return ast.WalkContinue, nil
}

type contentFilterExtension struct{}

// ExtContentFilter is a Goldmark extension that filters text content based on regex patterns.
var ExtContentFilter = &contentFilterExtension{}

// Extend adds the content filter extension to the provided Goldmark markdown processor.
func (e *contentFilterExtension) Extend(m goldmark.Markdown, filter *Filter) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&contentFilterRenderer{filter: filter}, 500),
	))
}
