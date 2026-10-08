// Package markdown — gno-button extension.
//
// `<gno-button href="…" label="…" variant="…" />` renders a link styled as a
// button. The tag is self-closing only; anything else falls through to raw
// HTML, which safe mode strips. The parser emits a plain *ast.Link, so the
// link extension owns the href (resolution, rel, icons, dangerous-URL guard).
package markdown

import (
	"bytes"
	"slices"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	mdhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"golang.org/x/net/html"
)

const buttonClassBase = "gno-button"

// buttonVariants is the whitelist of `variant` values, in the order their
// classes are emitted. Values combine (`variant="caution outline"`), match
// case-insensitively like alert kinds, and anything not listed is ignored.
var buttonVariants = []string{"outline", "caution", "warning", "info", "note"}

// buttonTagPrefix is matched case-insensitively, as x/net/html lowercases
// tag names (same as the sibling gno-* tags, see parseForeignLineTag).
var buttonTagPrefix = []byte("<gno-button")

// parseButtonTag reads a self-closing <gno-button … /> at the start of b
// and returns the token and the number of bytes it spans. Anything else —
// another tag, a start tag without `/>`, an unterminated tag — is not ok.
// The prefix check runs first so every other '<' costs no allocation.
func parseButtonTag(b []byte) (tok html.Token, n int, ok bool) {
	p := len(buttonTagPrefix)
	if len(b) <= p || !bytes.EqualFold(b[:p], buttonTagPrefix) {
		return tok, 0, false
	}
	// The tag name must end here: `<gno-buttons …/>` is another tag.
	if c := b[p]; c != '/' && !util.IsSpace(c) {
		return tok, 0, false
	}

	z := html.NewTokenizer(bytes.NewReader(b))
	if z.Next() != html.SelfClosingTagToken {
		return tok, 0, false
	}
	n = len(z.Raw()) // read before Token(), which may reuse the buffer
	return z.Token(), n, true
}

// newButtonLink builds the link node for a button tag, or returns nil when
// a required attribute is missing or the href is not allowed.
func newButtonLink(tok html.Token) *ast.Link {
	href, _ := ExtractAttr(tok.Attr, "href")
	label, _ := ExtractAttr(tok.Attr, "label")
	href, label = strings.TrimSpace(href), strings.TrimSpace(label)
	if href == "" || label == "" || !isButtonHrefAllowed(href) {
		return nil
	}

	variant, _ := ExtractAttr(tok.Attr, "variant")

	link := ast.NewLink()
	link.Destination = []byte(href)
	link.SetAttribute(linkClassAttr, buttonClass(strings.Fields(strings.ToLower(variant))))

	// Raw: the label is escaped on output but not re-parsed as markdown.
	labelNode := ast.NewString([]byte(label))
	labelNode.SetRaw(true)
	link.AppendChild(link, labelNode)
	return link
}

// isButtonHrefAllowed rejects what renderGnoLink would neutralize anyway
// (javascript:, vbscript:, file:), every data: URI, which goldmark allows
// for images but has no business behind a button, and any control byte:
// browsers strip tab and newline from a URL, so `java&#x09;script:` is a
// scheme the prefix check cannot see. It checks the resolved bytes the
// renderer will emit, see resolveDestination.
func isButtonHrefAllowed(href string) bool {
	dest := trimLeadingControlAndSpace(resolveDestination([]byte(href)))
	if bytes.ContainsFunc(dest, func(r rune) bool { return r < ' ' || r == 0x7f }) {
		return false
	}
	return !mdhtml.IsDangerousURL(dest) && !(len(dest) >= 5 && bytes.EqualFold(dest[:5], []byte("data:")))
}

// buttonClass returns the class attribute for the given variant words.
func buttonClass(words []string) string {
	class := buttonClassBase
	for _, v := range buttonVariants {
		if slices.Contains(words, v) {
			class += " " + buttonClassBase + "-" + v
		}
	}
	return class
}

// buttonParser is the inline parser for <gno-button … />.
type buttonParser struct{}

var _ parser.InlineParser = (*buttonParser)(nil)

func (*buttonParser) Trigger() []byte { return []byte{'<'} }

func (*buttonParser) Parse(_ ast.Node, block text.Reader, pc parser.Context) ast.Node {
	// A button inside a link label would nest <a> elements.
	if pc.IsInLinkLabel() {
		return nil
	}

	line, _ := block.PeekLine()
	tok, n, ok := parseButtonTag(line)
	if !ok {
		return nil
	}
	link := newButtonLink(tok)
	if link == nil {
		return nil
	}

	block.Advance(n)
	return link
}

// buttonBlockParser keeps a line holding only a button tag out of
// CommonMark's type-7 HTML block, which would otherwise take the line (and
// every line up to the next blank one) before inline parsing runs. It opens
// such a line as an ordinary paragraph; everything else is delegated to
// goldmark's paragraph parser, so the result behaves exactly like one.
type buttonBlockParser struct {
	parser.BlockParser
}

var _ parser.BlockParser = (*buttonBlockParser)(nil)

func (*buttonBlockParser) Trigger() []byte { return []byte{'<'} }

func (p *buttonBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	line = util.TrimRightSpace(util.TrimLeftSpace(line))
	if _, n, ok := parseButtonTag(line); !ok || n != len(line) {
		return nil, parser.NoChildren
	}
	return p.BlockParser.Open(parent, reader, pc)
}

type buttonExtension struct{}

// ExtButtons is the gno-button extension instance.
var ExtButtons = &buttonExtension{}

// Extend registers the button parsers. Priorities sit just ahead of the
// goldmark parsers they must pre-empt: raw HTML inline (400) and the HTML
// block (900). Rendering goes through the link extension.
func (e *buttonExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&buttonParser{}, 399),
		),
		parser.WithBlockParsers(
			util.Prioritized(&buttonBlockParser{parser.NewParagraphParser()}, 899),
		),
	)
}
