// gno-button extension.
//
// `<gno-button href="…" label="…" variant="…" />` renders a link styled as a
// button. The tag is self-closing only; anything else falls through to raw
// HTML, which safe mode strips. The parser emits a plain *ast.Link, so the
// link extension owns the href (resolution, rel, icons, dangerous-URL guard).

package markdown

import (
	"bytes"
	"html"
	"slices"
	"strings"
	"unicode"

	chainmd "github.com/gnolang/gno/gnovm/stdlibs/chain/markdown"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	mdhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

const buttonClassBase = "gno-button"

// buttonVariants is the whitelist of `variant` values, in the order their
// classes are emitted. Values combine (`variant="caution outline"`), match
// case-insensitively like alert kinds, and anything not listed is ignored.
var buttonVariants = []string{"outline", "caution", "warning", "info", "note"}

var buttonTagPrefix = []byte("<gno-button")

// maxButtonTagLen bounds how far a tag is scanned (see scanGnoTag), so a line
// of unterminated tags stays linear; 2 KB leaves room for a long href.
const maxButtonTagLen = 2048

// buttonTag is what a tag says. Values are raw attribute bytes aliasing the
// source; the first occurrence of an attribute wins, as in HTML.
type buttonTag struct {
	href, label, variant []byte
}

// parseButtonTag reads a self-closing `<gno-button … />` at the start of src,
// without allocating. It returns the tag's length, or 0 when src does not
// start with one (a tag ending in `>` is not a button).
func parseButtonTag(src []byte) (size int, t buttonTag) {
	var hasHref, hasLabel, hasVariant bool
	size, selfClosing := scanGnoTag(src, buttonTagPrefix, maxButtonTagLen, func(key, val []byte) {
		switch {
		case !hasHref && bytes.EqualFold(key, []byte("href")):
			t.href, hasHref = bytes.TrimSpace(val), true
		case !hasLabel && bytes.EqualFold(key, []byte("label")):
			t.label, hasLabel = val, true
		case !hasVariant && bytes.EqualFold(key, []byte("variant")):
			t.variant, hasVariant = val, true
		}
	})
	if !selfClosing {
		return 0, buttonTag{}
	}
	return size, t
}

// newButtonLink builds the link node for a button tag, or returns nil when
// a required attribute is missing or the href is not allowed.
func newButtonLink(t buttonTag) *ast.Link {
	label := buttonLabel(t.label)
	if len(t.href) == 0 || label == "" || !isButtonHrefAllowed(t.href) {
		return nil
	}

	// The raw href goes to the link pipeline, whose resolveDestination
	// decodes it once, exactly as it decodes a markdown link destination.
	link := ast.NewLink()
	link.Destination = bytes.Clone(t.href)
	link.SetAttribute(linkClassAttr, buttonClass(strings.Fields(strings.ToLower(string(t.variant)))))

	// Escaped on output, never parsed as markdown (raw).
	labelNode := ast.NewString([]byte(label))
	labelNode.SetRaw(true)
	link.AppendChild(link, labelNode)
	return link
}

// buttonLabel turns the raw label attribute into the text the button shows.
// Entities are decoded as HTML decodes attribute text. A button looks like
// first-party chrome, so what could make it read other than it renders is
// removed: bidi and zero-width characters (the set sanitize strips), and
// control characters, a line break or tab becoming a space. Trimming comes
// last, so `&#32;` or `&#x200B;` is no label.
func buttonLabel(raw []byte) string {
	label := chainmd.StripBidiAndZeroWidth(html.UnescapeString(string(raw)))
	label = strings.Map(func(r rune) rune {
		switch {
		case !unicode.IsControl(r):
			return r
		case unicode.IsSpace(r):
			return ' '
		}
		return -1
	}, label)
	return strings.TrimSpace(label)
}

// isButtonHrefAllowed rejects what renderGnoLink would neutralize anyway
// (javascript:, vbscript:, file:), every data: URI, which goldmark allows
// for images but has no business behind a button, and any control byte:
// browsers strip tab and newline from a URL, so `java&#x09;script:` is a
// scheme the prefix check cannot see. It checks the resolved bytes the
// renderer will emit, see resolveDestination.
func isButtonHrefAllowed(href []byte) bool {
	dest := trimLeadingControlAndSpace(resolveDestination(href))
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

// A button inside a link label needs no guard: goldmark does not build a link
// whose label holds a link (CommonMark: the inner one wins), so the brackets
// stay text, and an image alt renders its children as text only.
func (*buttonParser) Parse(parent ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, _ := block.PeekLine()
	n, tag := parseButtonTag(line)
	if n == 0 {
		return nil
	}
	link := newButtonLink(tag)
	if link == nil {
		return nil
	}

	block.Advance(n)
	// An alert title renders as <summary>, which must not hold a link
	// (interactive content): the button is reduced to its label text. Plain
	// markdown links in a title still render as links, as on master; that
	// belongs to the alert extension, not here.
	if inAlertHeader(parent) {
		label := link.FirstChild()
		link.RemoveChild(link, label)
		return label
	}
	return link
}

// inAlertHeader reports whether the block being inline-parsed is an alert
// title: the alert header holds it directly (see alertHeaderParser.Open).
func inAlertHeader(block ast.Node) bool {
	p := block.Parent()
	return p != nil && p.Kind() == KindAlertHeader
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
			util.Prioritized(newGnoTagLineParser(buttonTagPrefix), 899),
		),
	)
}
