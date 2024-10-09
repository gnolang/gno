package varmeta_md

import (
	"bytes"
	"strings"
	"unicode/utf8"

	"gno.land/p/demo/ufmt"
)

const VERSION = "1.5"

const (
	EXTENSION_NO_INTRA_EMPHASIS          = 1 << iota // ignore emphasis markers inside words
	EXTENSION_TABLES                                 // render tables
	EXTENSION_FENCED_CODE                            // render fenced code blocks
	EXTENSION_AUTOLINK                               // detect embedded URLs that are not explicitly marked
	EXTENSION_STRIKETHROUGH                          // strikethrough text using ~~test~~
	EXTENSION_LAX_HTML_BLOCKS                        // loosen up HTML block parsing rules
	EXTENSION_SPACE_HEADERS                          // be strict about prefix header rules
	EXTENSION_HARD_LINE_BREAK                        // translate newlines into line breaks
	EXTENSION_TAB_SIZE_EIGHT                         // expand tabs to eight spaces instead of four
	EXTENSION_FOOTNOTES                              // Pandoc-style footnotes
	EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK             // No need to insert an empty line to start a (code, quote, ordered list, unordered list) block
	EXTENSION_HEADER_IDS                             // specify header IDs  with {#id}
	EXTENSION_TITLEBLOCK                             // Titleblock ala pandoc
	EXTENSION_AUTO_HEADER_IDS                        // Create the header ID from the text
	EXTENSION_BACKSLASH_LINE_BREAK                   // translate trailing backslashes into line breaks
	EXTENSION_DEFINITION_LISTS                       // render definition lists
	EXTENSION_JOIN_LINES                             // delete newline and join lines

	commonHtmlFlags = 0 |
		HTML_USE_XHTML

	commonExtensions = 0 |
		EXTENSION_NO_INTRA_EMPHASIS |
		EXTENSION_TABLES |
		EXTENSION_FENCED_CODE |
		EXTENSION_AUTOLINK |
		EXTENSION_STRIKETHROUGH |
		EXTENSION_SPACE_HEADERS |
		EXTENSION_HEADER_IDS |
		EXTENSION_BACKSLASH_LINE_BREAK |
		EXTENSION_DEFINITION_LISTS
)

const (
	LINK_TYPE_NOT_AUTOLINK = iota
	LINK_TYPE_NORMAL
	LINK_TYPE_EMAIL
)

const (
	LIST_TYPE_ORDERED = 1 << iota
	LIST_TYPE_DEFINITION
	LIST_TYPE_TERM
	LIST_ITEM_CONTAINS_BLOCK
	LIST_ITEM_BEGINNING_OF_LIST
	LIST_ITEM_END_OF_LIST
)

const (
	TABLE_ALIGNMENT_LEFT = 1 << iota
	TABLE_ALIGNMENT_RIGHT
	TABLE_ALIGNMENT_CENTER = (TABLE_ALIGNMENT_LEFT | TABLE_ALIGNMENT_RIGHT)
)

const (
	TAB_SIZE_DEFAULT = 4
	TAB_SIZE_EIGHT   = 8
)

var blockTags = map[string]struct{}{
	"blockquote": {},
	"del":        {},
	"div":        {},
	"dl":         {},
	"fieldset":   {},
	"form":       {},
	"h1":         {},
	"h2":         {},
	"h3":         {},
	"h4":         {},
	"h5":         {},
	"h6":         {},
	"iframe":     {},
	"ins":        {},
	"math":       {},
	"noscript":   {},
	"ol":         {},
	"pre":        {},
	"p":          {},
	"script":     {},
	"style":      {},
	"table":      {},
	"ul":         {},

	"address":    {},
	"article":    {},
	"aside":      {},
	"canvas":     {},
	"details":    {},
	"figcaption": {},
	"figure":     {},
	"footer":     {},
	"header":     {},
	"hgroup":     {},
	"main":       {},
	"nav":        {},
	"output":     {},
	"progress":   {},
	"section":    {},
	"summary":    {},
	"video":      {},
}

type Renderer interface {
	BlockCode(out *bytes.Buffer, text []byte, infoString string)
	BlockQuote(out *bytes.Buffer, text []byte)
	BlockHtml(out *bytes.Buffer, text []byte)
	Header(out *bytes.Buffer, text func() bool, level int, id string)
	HRule(out *bytes.Buffer)
	List(out *bytes.Buffer, text func() bool, flags int)
	ListItem(out *bytes.Buffer, text []byte, flags int)
	Paragraph(out *bytes.Buffer, text func() bool)
	Table(out *bytes.Buffer, header []byte, body []byte, columnData []int)
	TableRow(out *bytes.Buffer, text []byte)
	TableHeaderCell(out *bytes.Buffer, text []byte, flags int)
	TableCell(out *bytes.Buffer, text []byte, flags int)
	Footnotes(out *bytes.Buffer, text func() bool)
	FootnoteItem(out *bytes.Buffer, name, text []byte, flags int)
	TitleBlock(out *bytes.Buffer, text []byte)

	AutoLink(out *bytes.Buffer, link []byte, kind int)
	CodeSpan(out *bytes.Buffer, text []byte)
	DoubleEmphasis(out *bytes.Buffer, text []byte)
	Emphasis(out *bytes.Buffer, text []byte)
	Image(out *bytes.Buffer, link []byte, title []byte, alt []byte)
	LineBreak(out *bytes.Buffer)
	Link(out *bytes.Buffer, link []byte, title []byte, content []byte)
	RawHtmlTag(out *bytes.Buffer, tag []byte)
	TripleEmphasis(out *bytes.Buffer, text []byte)
	StrikeThrough(out *bytes.Buffer, text []byte)
	FootnoteRef(out *bytes.Buffer, ref []byte, id int)

	Entity(out *bytes.Buffer, entity []byte)
	NormalText(out *bytes.Buffer, text []byte)

	DocumentHeader(out *bytes.Buffer)
	DocumentFooter(out *bytes.Buffer)

	GetFlags() int
}

type inlineParser func(p *parser, out *bytes.Buffer, data []byte, offset int) int

type parser struct {
	r              Renderer
	refOverride    ReferenceOverrideFunc
	refs           map[string]*reference
	inlineCallback [256]inlineParser
	flags          int
	nesting        int
	maxNesting     int
	insideLink     bool

	notes       []*reference
	notesRecord map[string]struct{}
}

func (p *parser) getRef(refid string) (ref *reference, found bool) {
	if p.refOverride != nil {
		r, overridden := p.refOverride(refid)
		if overridden {
			if r == nil {
				return nil, false
			}
			return &reference{
				link:     []byte(r.Link),
				title:    []byte(r.Title),
				noteId:   0,
				hasBlock: false,
				text:     []byte(r.Text),
			}, true
		}
	}
	ref, found = p.refs[strings.ToLower(refid)]
	return ref, found
}

func (p *parser) isFootnote(ref *reference) bool {
	_, ok := p.notesRecord[string(ref.link)]
	return ok
}

type Reference struct {
	Link  string
	Title string
	Text  string
}

type ReferenceOverrideFunc func(reference string) (ref *Reference, overridden bool)

type Options struct {
	Extensions        int
	ReferenceOverride ReferenceOverrideFunc
}

func MarkdownBasic(input []byte) []byte {
	htmlFlags := HTML_USE_XHTML
	renderer := HtmlRenderer(htmlFlags, "", "")

	return MarkdownOptions(input, renderer, Options{Extensions: 0})
}

func MarkdownCommon(input []byte) []byte {
	renderer := HtmlRenderer(commonHtmlFlags, "", "")
	return MarkdownOptions(input, renderer, Options{
		Extensions: commonExtensions,
	})
}

func Markdown(input []byte, renderer Renderer, extensions int) []byte {
	return MarkdownOptions(input, renderer, Options{
		Extensions: extensions,
	})
}

func MarkdownOptions(input []byte, renderer Renderer, opts Options) []byte {
	if renderer == nil {
		return nil
	}

	extensions := opts.Extensions

	p := new(parser)
	p.r = renderer
	p.flags = extensions
	p.refOverride = opts.ReferenceOverride
	p.refs = make(map[string]*reference)
	p.maxNesting = 16
	p.insideLink = false

	p.inlineCallback['*'] = emphasis
	p.inlineCallback['_'] = emphasis
	if extensions&EXTENSION_STRIKETHROUGH != 0 {
		p.inlineCallback['~'] = emphasis
	}
	p.inlineCallback['`'] = codeSpan
	p.inlineCallback['\n'] = lineBreak
	p.inlineCallback['['] = link
	p.inlineCallback['<'] = leftAngle
	p.inlineCallback['\\'] = escape
	p.inlineCallback['&'] = entity

	if extensions&EXTENSION_AUTOLINK != 0 {
		p.inlineCallback[':'] = autoLink
	}

	if extensions&EXTENSION_FOOTNOTES != 0 {
		p.notes = make([]*reference, 0)
		p.notesRecord = make(map[string]struct{})
	}

	first := firstPass(p, input)
	second := secondPass(p, first)
	return second
}

func firstPass(p *parser, input []byte) []byte {
	var out bytes.Buffer
	tabSize := TAB_SIZE_DEFAULT
	if p.flags&EXTENSION_TAB_SIZE_EIGHT != 0 {
		tabSize = TAB_SIZE_EIGHT
	}
	beg := 0
	lastFencedCodeBlockEnd := 0
	for beg < len(input) {
		end := beg
		for end < len(input) && input[end] != '\n' && input[end] != '\r' {
			end++
		}

		if p.flags&EXTENSION_FENCED_CODE != 0 {
			if beg >= lastFencedCodeBlockEnd {
				if i := p.fencedCodeBlock(&out, input[beg:], false); i > 0 {
					lastFencedCodeBlockEnd = beg + i
				}
			}
		}

		if end > beg {
			if end < lastFencedCodeBlockEnd {
				out.Write(input[beg:end])
			} else if refEnd := isReference(p, input[beg:], tabSize); refEnd > 0 {
				beg += refEnd
				continue
			} else {
				expandTabs(&out, input[beg:end], tabSize)
			}
		}

		if end < len(input) && input[end] == '\r' {
			end++
		}
		if end < len(input) && input[end] == '\n' {
			end++
		}
		out.WriteByte('\n')

		beg = end
	}

	if out.Len() == 0 {
		out.WriteByte('\n')
	}

	return out.Bytes()
}

func secondPass(p *parser, input []byte) []byte {
	var output bytes.Buffer

	p.r.DocumentHeader(&output)
	p.block(&output, input)

	if p.flags&EXTENSION_FOOTNOTES != 0 && len(p.notes) > 0 {
		p.r.Footnotes(&output, func() bool {
			flags := LIST_ITEM_BEGINNING_OF_LIST
			for i := 0; i < len(p.notes); i += 1 {
				ref := p.notes[i]
				var buf bytes.Buffer
				if ref.hasBlock {
					flags |= LIST_ITEM_CONTAINS_BLOCK
					p.block(&buf, ref.title)
				} else {
					p.inline(&buf, ref.title)
				}
				p.r.FootnoteItem(&output, ref.link, buf.Bytes(), flags)
				flags &^= LIST_ITEM_BEGINNING_OF_LIST | LIST_ITEM_CONTAINS_BLOCK
			}

			return true
		})
	}

	p.r.DocumentFooter(&output)

	if p.nesting != 0 {
		panic("Nesting level did not end at zero")
	}

	return output.Bytes()
}

type reference struct {
	link     []byte
	title    []byte
	noteId   int
	hasBlock bool
	text     []byte
}

func (r *reference) String() string {
	return ufmt.Sprintf("{link: %q, title: %q, text: %q, noteId: %d, hasBlock: %v}",
		r.link, r.title, r.text, r.noteId, r.hasBlock)
}

func isReference(p *parser, data []byte, tabSize int) int {
	if len(data) < 4 {
		return 0
	}
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}

	noteId := 0

	if data[i] != '[' {
		return 0
	}
	i++
	if p.flags&EXTENSION_FOOTNOTES != 0 {
		if i < len(data) && data[i] == '^' {
			noteId = 1
			i++
		}
	}
	idOffset := i
	for i < len(data) && data[i] != '\n' && data[i] != '\r' && data[i] != ']' {
		i++
	}
	if i >= len(data) || data[i] != ']' {
		return 0
	}
	idEnd := i

	i++
	if i >= len(data) || data[i] != ':' {
		return 0
	}
	i++
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i < len(data) && (data[i] == '\n' || data[i] == '\r') {
		i++
		if i < len(data) && data[i] == '\n' && data[i-1] == '\r' {
			i++
		}
	}
	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i >= len(data) {
		return 0
	}

	var (
		linkOffset, linkEnd   int
		titleOffset, titleEnd int
		lineEnd               int
		raw                   []byte
		hasBlock              bool
	)

	if p.flags&EXTENSION_FOOTNOTES != 0 && noteId != 0 {
		linkOffset, linkEnd, raw, hasBlock = scanFootnote(p, data, i, tabSize)
		lineEnd = linkEnd
	} else {
		linkOffset, linkEnd, titleOffset, titleEnd, lineEnd = scanLinkRef(p, data, i)
	}
	if lineEnd == 0 {
		return 0
	}

	ref := &reference{
		noteId:   noteId,
		hasBlock: hasBlock,
	}

	if noteId > 0 {
		ref.link = data[idOffset:idEnd]
		ref.title = raw
	} else {
		ref.link = data[linkOffset:linkEnd]
		ref.title = data[titleOffset:titleEnd]
	}

	id := string(bytes.ToLower(data[idOffset:idEnd]))

	p.refs[id] = ref

	return lineEnd
}

func scanLinkRef(p *parser, data []byte, i int) (linkOffset, linkEnd, titleOffset, titleEnd, lineEnd int) {
	if data[i] == '<' {
		i++
	}
	linkOffset = i
	if i == len(data) {
		return
	}
	for i < len(data) && data[i] != ' ' && data[i] != '\t' && data[i] != '\n' && data[i] != '\r' {
		i++
	}
	linkEnd = i
	if data[linkOffset] == '<' && data[linkEnd-1] == '>' {
		linkOffset++
		linkEnd--
	}

	for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
		i++
	}
	if i < len(data) && data[i] != '\n' && data[i] != '\r' && data[i] != '\'' && data[i] != '"' && data[i] != '(' {
		return
	}

	if i >= len(data) || data[i] == '\r' || data[i] == '\n' {
		lineEnd = i
	}
	if i+1 < len(data) && data[i] == '\r' && data[i+1] == '\n' {
		lineEnd++
	}

	if lineEnd > 0 {
		i = lineEnd + 1
		for i < len(data) && (data[i] == ' ' || data[i] == '\t') {
			i++
		}
	}

	if i+1 < len(data) && (data[i] == '\'' || data[i] == '"' || data[i] == '(') {
		i++
		titleOffset = i

		for i < len(data) && data[i] != '\n' && data[i] != '\r' {
			i++
		}
		if i+1 < len(data) && data[i] == '\n' && data[i+1] == '\r' {
			titleEnd = i + 1
		} else {
			titleEnd = i
		}

		i--
		for i > titleOffset && (data[i] == ' ' || data[i] == '\t') {
			i--
		}
		if i > titleOffset && (data[i] == '\'' || data[i] == '"' || data[i] == ')') {
			lineEnd = titleEnd
			titleEnd = i
		}
	}

	return
}

func scanFootnote(p *parser, data []byte, i, indentSize int) (blockStart, blockEnd int, contents []byte, hasBlock bool) {
	if i == 0 || len(data) == 0 {
		return
	}

	for i < len(data) && data[i] == ' ' {
		i++
	}

	blockStart = i

	blockEnd = i
	for i < len(data) && data[i-1] != '\n' {
		i++
	}

	var raw bytes.Buffer

	raw.Write(data[blockEnd:i])
	blockEnd = i

	containsBlankLine := false

gatherLines:
	for blockEnd < len(data) {
		i++

		for i < len(data) && data[i-1] != '\n' {
			i++
		}

		if p.isEmpty(data[blockEnd:i]) > 0 {
			containsBlankLine = true
			blockEnd = i
			continue
		}

		n := 0
		if n = isIndented(data[blockEnd:i], indentSize); n == 0 {
			break gatherLines
		}

		if containsBlankLine {
			raw.WriteByte('\n')
			containsBlankLine = false
		}

		raw.Write(data[blockEnd+n : i])
		hasBlock = true

		blockEnd = i
	}

	if data[blockEnd-1] != '\n' {
		raw.WriteByte('\n')
	}

	contents = raw.Bytes()

	return
}

func ispunct(c byte) bool {
	for _, r := range []byte("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~") {
		if c == r {
			return true
		}
	}
	return false
}

func isspace(c byte) bool {
	return ishorizontalspace(c) || isverticalspace(c)
}

func ishorizontalspace(c byte) bool {
	return c == ' ' || c == '\t'
}

func isverticalspace(c byte) bool {
	return c == '\n' || c == '\r' || c == '\f' || c == '\v'
}

func isletter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isalnum(c byte) bool {
	return (c >= '0' && c <= '9') || isletter(c)
}

func expandTabs(out *bytes.Buffer, line []byte, tabSize int) {
	i, prefix := 0, 0
	slowcase := false
	for i = 0; i < len(line); i++ {
		if line[i] == '\t' {
			if prefix == i {
				prefix++
			} else {
				slowcase = true
				break
			}
		}
	}

	if !slowcase {
		for i = 0; i < prefix*tabSize; i++ {
			out.WriteByte(' ')
		}
		out.Write(line[prefix:])
		return
	}

	column := 0
	i = 0
	for i < len(line) {
		start := i
		for i < len(line) && line[i] != '\t' {
			_, size := utf8.DecodeRune(line[i:])
			i += size
			column++
		}

		if i > start {
			out.Write(line[start:i])
		}

		if i >= len(line) {
			break
		}

		for {
			out.WriteByte(' ')
			column++
			if column%tabSize == 0 {
				break
			}
		}

		i++
	}
}

func isIndented(data []byte, indentSize int) int {
	if len(data) == 0 {
		return 0
	}
	if data[0] == '\t' {
		return 1
	}
	if len(data) < indentSize {
		return 0
	}
	for i := 0; i < indentSize; i++ {
		if data[i] != ' ' {
			return 0
		}
	}
	return indentSize
}

func slugify(in []byte) []byte {
	if len(in) == 0 {
		return in
	}
	out := make([]byte, 0, len(in))
	sym := false

	for _, ch := range in {
		if isalnum(ch) {
			sym = false
			out = append(out, ch)
		} else if sym {
			continue
		} else {
			out = append(out, '-')
			sym = true
		}
	}
	var a, b int
	var ch byte
	for a, ch = range out {
		if ch != '-' {
			break
		}
	}
	for b = len(out) - 1; b > 0; b-- {
		if out[b] != '-' {
			break
		}
	}
	return out[a : b+1]
}
