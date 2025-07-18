package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"golang.org/x/net/html"

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

// Parse parses button syntax: {text}(url), {text|options}(url) and <gno-button href="url" content="..."/>
func (p *buttonParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, seg := block.PeekLine()
	if len(line) == 0 {
		return nil
	}

	// Parse {text}(url) or {text|options}(url) syntax
	if line[0] == '{' {
		return p.parseCurlyButton(line, seg, block)
	}

	// Parse <gno-button href="url" content="..."/> syntax
	trimmedLine := bytes.TrimSpace(line)
	if bytes.HasPrefix(trimmedLine, []byte("<gno-button")) {
		return p.parseGnoButton(trimmedLine, seg, block)
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
	options := p.ExtractOptions(parts)

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

// parseGnoButton parses <gno-button href="url" content="..." variant="..."/> syntax
func (p *buttonParser) parseGnoButton(line []byte, seg text.Segment, block text.Reader) ast.Node {
    toks, err := ParseHTMLTokens(bytes.NewReader(line))
    if err != nil || len(toks) != 1 {
        return nil
    }
    tok := toks[0]

    // We only want self-closing tags <gno-button … />
    if tok.Type != html.SelfClosingTagToken || tok.Data != "gno-button" {
        return nil
    }

    // Extract the href attribute and content (mandatory attributes)
    href, ok := ExtractAttr(tok.Attr, "href")
    if !ok {
        return nil
    }
    label, ok := ExtractAttr(tok.Attr, "content")
    if !ok {
        return nil
    }
    variantStr, _ := ExtractAttr(tok.Attr, "variant")

    // Validate the variants
    options := []string{}
    for _, opt := range strings.Fields(variantStr) {
        if allowedButtonOptions[opt] {
            options = append(options, opt)
        }
    }

    // Create the button link
    link := p.createButtonLink(href, options)

    // Create and add the text node with the button label
    textNode := ast.NewText()
    idx := max(0, bytes.Index(line, []byte(label)))
    startOffset := seg.Start + idx
    textNode.Segment = text.Segment{Start: startOffset, Stop: startOffset + len(label)}
    link.AppendChild(link, textNode)

    // Advance the reader
    block.Advance(len(line))
    return link
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

type buttonExtension struct{}

var _ goldmark.Extender = (*buttonExtension)(nil)

// Extend registers the button parsers
func (e *buttonExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(NewButtonParser(), 200)),
	)
}

// ExtButtons is the exported extension instance
var ExtButtons = &buttonExtension{}
