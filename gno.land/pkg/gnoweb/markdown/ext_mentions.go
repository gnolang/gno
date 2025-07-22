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

// isInsideMarkdownLink checks if the current position is inside a Markdown link
func isInsideMarkdownLink(source []byte, pos int) bool {
	for i := pos - 1; i >= 0; i-- {
		if source[i] == '\n' {
			return false
		}
		if source[i] == '[' {
			for j := pos; j < len(source); j++ {
				if source[j] == '\n' {
					return false
				}
				if source[j] == ']' && j+1 < len(source) && source[j+1] == '(' {
					return true
				}
			}
			return false
		}
	}
	return false
}

// findG1Address finds a valid g1 address in the given text starting at the specified position
func getG1Address(text []byte, startPos int) (string, int, bool) {
	if len(text) < startPos+2 || text[startPos] != 'g' || text[startPos+1] != '1' {
		return "", 0, false
	}

	if m := gnoPattern.FindIndex(text[startPos:]); m != nil && m[0] == 0 {
		addr := string(text[startPos : startPos+m[1]])
		if isValidWordBoundary(text, startPos, startPos+m[1]) && IsValidBech32Address(addr) {
			return addr, m[1], true
		}
	}

	return "", 0, false
}

// tryFindG1Address tries to find a g1 address at the given position and creates a link if found
func findG1Address(line []byte, segStart int, linePos int, seg text.Segment, parent ast.Node, block text.Reader) ast.Node {
	if foundAddr, length, found := getG1Address(line, linePos); found {
		start := segStart
		stop := start + length
		advance := length
		var spaceNode ast.Node

		if linePos > 0 {
			// This is a space-prefixed address
			spaceNode = ast.NewTextSegment(text.NewSegment(seg.Start, seg.Start+1))
			start = segStart + 1
			advance = length + 1
		}

		// Create the link
		linkNode := createMentionLink(start, stop, "/u/"+foundAddr, nil)

		block.Advance(advance)

		if spaceNode != nil {
			parent.AppendChild(parent, spaceNode)
		}

		return linkNode
	}

	return nil
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
	return []byte{'@', ' ', 'g'}
}

// Parse parses both @ mentions and g1 addresses
// TODO: check if the user exists before creating the link
func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	_, seg := block.Position()

	if len(line) == 0 {
		return nil
	}

	// Handle g1 addresses
	// Check for g1 address (at start of line)
	if len(line) > 1 && line[0] == 'g' && line[1] == '1' {
		// Check if this is at the start of the line
		source := block.Source()
		isStartOfLine := seg.Start == 0 || source[seg.Start-1] == '\n'

		if isStartOfLine {
			if linkNode := findG1Address(line, seg.Start, 0, seg, parent, block); linkNode != nil {
				return linkNode
			}
		}
	}

	// Check for g1 address (preceded by space)
	if len(line) > 2 && line[0] == ' ' && line[1] == 'g' && line[2] == '1' {
		if linkNode := findG1Address(line, seg.Start, 1, seg, parent, block); linkNode != nil {
			return linkNode
		}
	}

	// Handle @ mentions
	if line[0] == '@' {
		// Check if we're inside a Markdown link
		if isInsideMarkdownLink(block.Source(), seg.Start) {
			return nil
		}

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
