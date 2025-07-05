package markdown

import (
	"bytes"
	"regexp"

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

// TODO: check is user exists
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

// mentionParser implements InlineParser for @ mentions
type mentionParser struct{}

// Trigger returns the bytes that trigger this parser
func (p *mentionParser) Trigger() []byte {
	return []byte{'@'}
}

// Parse parses @ mentions
func (p *mentionParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) == 0 || line[0] != '@' {
		return nil
	}

	// Avoid matching email addresses: skip if preceding character is alphanumeric, underscore, hyphen, or dot
	_, seg := block.Position()
	if seg.Start > 0 {
		src := block.Source()
		prev := src[seg.Start-1]
		if (prev >= 'A' && prev <= 'Z') || (prev >= 'a' && prev <= 'z') ||
			(prev >= '0' && prev <= '9') || prev == '_' || prev == '-' || prev == '.' {
			return nil
		}
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

// AST transformer for GNO addresses
type gnoTransformer struct{}

func (t *gnoTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		txt, ok := n.(*ast.Text)
		if !ok || !entering {
			return ast.WalkContinue, nil
		}
		seg := txt.Segment
		raw := source[seg.Start:seg.Stop]
		matches := gnoPattern.FindAllIndex(raw, -1)
		if len(matches) == 0 {
			return ast.WalkContinue, nil
		}

		// Replace this Text node with a series of nodes
		parent := txt.Parent()
		lastEnd := 0

		for _, m := range matches {
			if m[0] > lastEnd {
				parent.InsertBefore(parent, txt, ast.NewTextSegment(
					text.NewSegment(seg.Start+lastEnd, seg.Start+m[0]),
				))
			}

			addr := string(raw[m[0]:m[1]])
			if IsValidBech32Address(addr) {
				destination := "/u/" + addr
				link := createMentionLink(seg.Start+m[0], seg.Start+m[1], destination, nil)
				parent.InsertBefore(parent, txt, link)
			} else {
				// If invalid, keep as text
				parent.InsertBefore(parent, txt, ast.NewTextSegment(
					text.NewSegment(seg.Start+m[0], seg.Start+m[1]),
				))
			}

			lastEnd = m[1]
		}

		if lastEnd < len(raw) {
			parent.InsertBefore(parent, txt, ast.NewTextSegment(
				text.NewSegment(seg.Start+lastEnd, seg.Stop),
			))
		}

		parent.RemoveChild(parent, txt)
		return ast.WalkSkipChildren, nil
	})
}

// mentionExtension is a Goldmark extension that registers the mention parser.
type mentionExtension struct{}

// ExtMention is the exported extension instance.
var ExtMention = &mentionExtension{}

// Extend registers the inline parser and AST transformer with high priority.
func (e *mentionExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&mentionParser{}, 300),
		),
		parser.WithASTTransformers(
			util.Prioritized(&gnoTransformer{}, 300),
		),
	)
}
