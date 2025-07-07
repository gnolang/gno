package markdown

import (
	"bytes"
	"regexp"
	"unicode"

	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var (
	mentionPattern = regexp.MustCompile(`^@([A-Za-z0-9_]{3,})`)
	gnoPattern     = regexp.MustCompile(`g1[0-9A-Za-z]{37,}`)
)

// IsValidBech32Address validates if the given string is a valid bech32 address
func IsValidBech32Address(address string) bool {
	_, _, err := bech32.Decode(address)
	return err == nil
}

// isValidWordBoundary checks if the position is at a valid word boundary
func isValidWordBoundary(text []byte, start, end int) bool {
	// Check character before the match
	if start > 0 {
		prev := rune(text[start-1])
		if unicode.IsLetter(prev) || unicode.IsDigit(prev) || prev == '_' {
			return false
		}
	}

	// Check character after the match
	if end < len(text) {
		next := rune(text[end])
		if unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_' {
			return false
		}
	}

	return true
}

// isValidEmailBoundary checks if @ is not part of an email address
func isValidEmailBoundary(text []byte, start int) bool {
	if start == 0 {
		return true
	}

	prev := text[start-1]
	return !((prev >= 'A' && prev <= 'Z') || (prev >= 'a' && prev <= 'z') ||
		(prev >= '0' && prev <= '9') || prev == '_' || prev == '-' || prev == '.')
}

// createMentionLink creates a link node for mentions and addresses
func createMentionLink(start, stop int, destination string, block text.Reader) ast.Node {
	link := ast.NewLink()
	link.Destination = []byte(destination)
	textNode := ast.NewTextSegment(text.NewSegment(start, stop))
	link.AppendChild(link, textNode)

	if block != nil {
		block.Advance(stop - start)
	}

	return link
}

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

// Parse parses both @ mentions and g1 addresses
// TODO: check if the user exists before creating the link
func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	_, seg := block.Position()

	if len(line) == 0 {
		return nil
	}

	// Handle g1 addresses preceded by a space
	if line[0] == ' ' && len(line) > 2 && line[1] == 'g' && line[2] == '1' {
		sub := line[1:]
		if m := gnoPattern.FindIndex(sub); m != nil && m[0] == 0 {
			addr := string(sub[:m[1]])
			if isValidWordBoundary(sub, m[0], m[1]) && IsValidBech32Address(addr) {
				// To keep the space before the link
				spaceNode := ast.NewTextSegment(text.NewSegment(seg.Start, seg.Start+1))

				start := seg.Start + 1
				stop := start + m[1]

				linkNode := createMentionLink(start, stop, "/u/"+addr, nil)

				block.Advance(m[1] + 1)

				parent.AppendChild(parent, spaceNode)
				return linkNode
			}
		}
	}

	// Handle @ mentions
	if line[0] == '@' {

		// Check email boundary
		if !isValidEmailBoundary(block.Source(), seg.Start) {
			return nil
		}

		mentionMatch := mentionPattern.FindSubmatch(line)
		if mentionMatch == nil {
			return nil
		}

		mention := string(mentionMatch[0])
		destination := "/u/" + string(mentionMatch[1])

		// Check boundary
		mentionEnd := len(mention)
		if mentionEnd < len(line) {
			const validDelimiters = " \t\n\r.,:;!?()[]"
			next := rune(line[mentionEnd])
			if !bytes.ContainsRune([]byte(validDelimiters), next) {
				return nil
			}
		}

		return createMentionLink(seg.Start, seg.Start+len(mention), destination, block)
	}

	return nil
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
