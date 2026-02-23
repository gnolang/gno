package markdown

import (
	_ "embed"
	"html"
	"log"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

//go:embed contentfilter_patterns.txt
var DefaultContentFilterPatterns string

// DefaultContentFilter is the pre-compiled content filter using embedded patterns.
var DefaultContentFilter = NewFilter(DefaultContentFilterPatterns)

// Filter provides text content filtering for Markdown rendering.
// Patterns are immutable after creation and safe for concurrent use.
type Filter struct {
	patterns           []*compiledPattern
	defaultReplacement string
}

type compiledPattern struct {
	regex       *regexp.Regexp
	replacement string
}

// NewFilter creates a content filter by parsing pattern definitions from content.
// Each line is a regex pattern, optionally followed by " -> " and a replacement.
// Lines starting with # are comments. Empty lines are ignored.
// Use DEFAULT_REPLACEMENT=... to set the fallback replacement text.
// Invalid patterns are logged and skipped.
func NewFilter(content string) *Filter {
	f := &Filter{
		defaultReplacement: "[filtered]",
	}

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "DEFAULT_REPLACEMENT=") {
			f.defaultReplacement = strings.TrimPrefix(line, "DEFAULT_REPLACEMENT=")
			continue
		}

		var patternStr, replacement string
		if idx := strings.Index(line, " -> "); idx != -1 {
			patternStr = strings.TrimSpace(line[:idx])
			replacement = strings.TrimSpace(line[idx+len(" -> "):])
		} else {
			patternStr = line
			replacement = ""
		}

		regex, err := regexp.Compile(patternStr)
		if err != nil {
			log.Printf("warning: invalid content filter pattern skipped: %q: %v", patternStr, err)
			continue
		}

		f.patterns = append(f.patterns, &compiledPattern{
			regex:       regex,
			replacement: replacement,
		})
	}

	return f
}

// FilterText applies all loaded patterns to text and returns the filtered result.
// Returns text unchanged if the Filter is nil.
func (f *Filter) FilterText(text string) string {
	if f == nil {
		return text
	}

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
