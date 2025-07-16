package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// buttonParser implements InlineParser for button syntax
type buttonParser struct{}

// allowedButtonOptions is a whitelist of allowed button variants
var allowedButtonOptions = map[string]bool{
	"outline": true,
	"caution": true,
	"warning": true,
	"info":    true,
	"note":    true,
}

// pipeRegex splits on pipe surrounded by any whitespace
var pipeRegex = regexp.MustCompile(`\s*\|\s*`)

var _ parser.InlineParser = (*buttonParser)(nil)

// NewButtonParser creates a new instance of buttonParser
func NewButtonParser() parser.InlineParser {
	return &buttonParser{}
}

// Trigger returns the bytes that trigger this parser
func (p *buttonParser) Trigger() []byte {
	return []byte{'{', '<'}
}

// Parse parses button syntax: {text}(url), {text|options}(url) and <gno-button href="url">text</gno-button>
func (p *buttonParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	// Parse {text}(url) or {text|options}(url) syntax
	if line[0] == '{' {
		return p.parseCurlyButton(line, seg, block)
	}

	// Parse <gno-button href="url">text</gno-button> syntax
	if bytes.HasPrefix(line, []byte("<gno-button")) {
		return p.parseGnoButton(line, seg, block)
	}

	return nil
}

// parseCurlyButton parses {text}(url) or {text|options}(url) syntax
func (p *buttonParser) parseCurlyButton(line []byte, seg text.Segment, block text.Reader) ast.Node {
	// Find the closing }(url) pattern
	closingIndex := bytes.Index(line, []byte("}("))
	if closingIndex == -1 {
		return nil
	}

	// Find the closing parenthesis
	parenIndex := bytes.Index(line[closingIndex:], []byte(")"))
	if parenIndex == -1 {
		return nil
	}

	// Extract URL and content
	url := string(line[closingIndex+2 : closingIndex+parenIndex])
	content := string(line[1:closingIndex])

	// Split content and extract label/options
	parts := pipeRegex.Split(content, 2)
	label := strings.TrimSpace(parts[0])
	options := p.extractOptions(parts)

	// Create link and text node
	link := p.createButtonLink(url, options)
	textNode := ast.NewText()

	// Calculate label position in original text
	labelStart := seg.Start + 1 + strings.Index(content, label)
	textNode.Segment = text.Segment{Start: labelStart, Stop: labelStart + len(label)}
	link.AppendChild(link, textNode)

	// Advance the reader
	block.Advance(closingIndex + parenIndex + 1)
	return link
}

// parseGnoButton parses <gno-button href="url" variant="...">text</gno-button> syntax
func (p *buttonParser) parseGnoButton(line []byte, seg text.Segment, block text.Reader) ast.Node {
	// Extract URL from href attribute
	hrefIndex := bytes.Index(line, []byte("href=\""))
	if hrefIndex == -1 {
		return nil
	}
	urlStart := hrefIndex + 6 // len('href="')
	urlEnd := bytes.Index(line[urlStart:], []byte("\""))
	if urlEnd == -1 {
		return nil
	}
	url := string(line[urlStart : urlStart+urlEnd])

	// Extract variant attribute if present
	options := []string{}
	variantIndex := bytes.Index(line, []byte("variant=\""))
	if variantIndex != -1 {
		variantStart := variantIndex + 9 // len('variant="')
		variantEnd := bytes.Index(line[variantStart:], []byte("\""))
		if variantEnd != -1 {
			variantValue := string(line[variantStart : variantStart+variantEnd])
			for _, opt := range strings.Fields(variantValue) {
				if allowedButtonOptions[opt] {
					options = append(options, opt)
				}
			}
		}
	}

	// Find text boundaries
	textStart := bytes.Index(line, []byte(">"))
	closeIndex := bytes.Index(line, []byte("</gno-button>"))
	if textStart == -1 || closeIndex == -1 {
		return nil
	}

	// Create link and text node
	link := p.createButtonLink(url, options)
	textNode := ast.NewText()
	textNode.Segment = text.Segment{
		Start: seg.Start + textStart + 1,
		Stop:  seg.Start + closeIndex,
	}
	link.AppendChild(link, textNode)

	// Advance the reader
	block.Advance(closeIndex + len([]byte("</gno-button>")))
	return link
}

// extractOptions extracts and validates options from the second part of split content
func (p *buttonParser) extractOptions(parts []string) []string {
	if len(parts) != 2 {
		return []string{}
	}

	options := []string{}
	for _, opt := range strings.Fields(parts[1]) {
		if allowedButtonOptions[opt] {
			options = append(options, opt)
		}
	}
	return options
}

// createButtonLink creates a link node with gno-button class and optional variants
func (p *buttonParser) createButtonLink(url string, options []string) *ast.Link {
	link := ast.NewLink()
	link.Destination = []byte(url)

	classes := []string{"gno-button"}
	for _, option := range options {
		classes = append(classes, "gno-button-"+option)
	}

	link.SetAttributeString("class", strings.Join(classes, " "))
	return link
}

// buttonExtension is the extension that adds button support
type buttonExtension struct{}

var _ goldmark.Extender = (*buttonExtension)(nil)

// Extend registers the button parsers
func (e *buttonExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(NewButtonParser(), 300)),
	)
}

// ExtButtons is the exported extension instance
var ExtButtons = &buttonExtension{}
