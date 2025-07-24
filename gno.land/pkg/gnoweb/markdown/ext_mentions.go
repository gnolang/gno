package markdown

import (
	"regexp"

	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// NewMentionParser creates a new parser for @ mentions and g1 addresses
func NewMentionParser() parser.InlineParser {
	return &mentionParser{}
}

// mentionParser implements InlineParser for both @ mentions and g1 addresses
type mentionParser struct{}

// Trigger returns the bytes that trigger this parser
func (p *mentionParser) Trigger() []byte {
	return []byte{'@', ' '}
}

var (
	mentionNamePattern    = regexp.MustCompile(`^@([A-Za-z0-9_]{3,90})(?:[^A-Za-z0-9_-]|$)`)
	mentionAddressPattern = regexp.MustCompile(`^(?: |)(g1[0-9A-Za-z]{37,90})(?:[^@0-9A-Za-z]|$)`)
)

// isValidBech32Address validates if the given string is a valid bech32 address
func isValidBech32Address(address []byte) bool {
	_, _, err := bech32.Decode(string(address))
	return err == nil
}

func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	current, before := rune(block.Peek()), block.PrecendingCharacter()

	line, lineSegment := block.PeekLine()
	if len(line) < 2 {
		return nil
	}

	if pc.IsInLinkLabel() { // if inside a link, skip
		return nil
	}

	switch current {
	case '@':
		if before != ' ' && before != '\n' {
			return nil // continue
		}

		match := mentionNamePattern.FindSubmatchIndex(line)
		if match == nil {
			return nil
		}
		start, end := match[2], match[3] // get first submatch

		mentionSeg := text.NewSegment(lineSegment.Start+start-1, lineSegment.Start+end)

		// Craft link node
		link := ast.NewLink()
		link.AppendChild(link, ast.NewTextSegment(mentionSeg))
		link.Destination = append([]byte("/u/"), line[start:end]...)

		block.Advance(end)

		return link

	case ' ', 'g':
		match := mentionAddressPattern.FindSubmatchIndex(line)
		if match == nil {
			return nil
		}
		start, end := match[2], match[3] // get first submatch

		if !isValidBech32Address(line[start:end]) { // validate address
			return nil
		}

		if current == ' ' {
			// Create a space before link
			space := ast.NewRawTextSegment(text.NewSegment(lineSegment.Start, lineSegment.Start+1))
			parent.AppendChild(parent, space)
		}

		mentionSeg := text.NewSegment(lineSegment.Start+start, lineSegment.Start+end)

		// Craft link node
		link := ast.NewLink()
		link.AppendChild(link, ast.NewTextSegment(mentionSeg))
		link.Destination = append([]byte("/u/"), line[start:end]...)

		block.Advance(end)

		return link

	default:
		return nil
	}
}

// mentionExtension is a Goldmark extension that registers the unified parser.
type mentionExtension struct{}

// ExtMention is the exported extension instance.
var ExtMention = &mentionExtension{}

// Extend registers the unified parser with high priority.
func (e *mentionExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(util.Prioritized(NewMentionParser(), 500)),
	)
}
