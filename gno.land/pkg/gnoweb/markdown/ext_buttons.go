package markdown

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// buttonParser implements InlineParser for button syntax
type buttonParser struct{}

var _ parser.InlineParser = (*buttonParser)(nil)

// NewButtonParser creates a new instance of buttonParser
func NewButtonParser() parser.InlineParser {
	return &buttonParser{}
}

// Trigger returns the bytes that trigger this parser
func (p *buttonParser) Trigger() []byte {
	return []byte{'{', '<'}
}

// Parse parses button syntax: {text}(url) and <gno-button href="url">text</gno-button>
func (p *buttonParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	// Parse {text}(url) syntax
	if line[0] == '{' {
		// Find the closing }(url) pattern
		closingPattern := []byte("}(")
		closingIndex := bytes.Index(line, closingPattern)
		if closingIndex == -1 {
			return nil
		}

		// Find the closing parenthesis
		parenIndex := bytes.Index(line[closingIndex:], []byte(")"))
		if parenIndex == -1 {
			return nil
		}

		// Extract URL and create link
		url := string(line[closingIndex+2 : closingIndex+parenIndex])
		link := p.createButtonLink(url)

		// Add text node with proper offsets
		_, seg := block.PeekLine()
		labelStart := seg.Start + 1 // position just after the opening '{'
		labelEnd := seg.Start + closingIndex

		textNode := ast.NewText()
		textNode.Segment = text.Segment{Start: labelStart, Stop: labelEnd}
		link.AppendChild(link, textNode)

		// Advance the reader
		block.Advance(closingIndex + parenIndex + 1)

		return link
	}

	// Parse <gno-button href="url">text</gno-button> syntax
	if bytes.HasPrefix(line, []byte("<gno-button")) {
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

		// Find the closing > and </gno-button>
		textStart := bytes.Index(line, []byte(">"))
		if textStart == -1 {
			return nil
		}
		closeIndex := bytes.Index(line, []byte("</gno-button>"))
		if closeIndex == -1 {
			return nil
		}

		// Create link
		link := p.createButtonLink(url)

		// Add text node with proper offsets
		_, seg := block.PeekLine()
		absStart := seg.Start + textStart + 1
		absEnd := seg.Start + closeIndex

		textNode := ast.NewText()
		textNode.Segment = text.Segment{Start: absStart, Stop: absEnd}
		link.AppendChild(link, textNode)

		// Advance the reader
		block.Advance(closeIndex + len([]byte("</gno-button>")))

		return link
	}

	return nil
}

// createButtonLink creates a link node with gno-button class
func (p *buttonParser) createButtonLink(url string) *ast.Link {
	link := ast.NewLink()
	link.Destination = []byte(url)
	link.SetAttributeString("class", "gno-button")
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
