package varmeta_md

import (
	"bytes"
	"regexp"
	"strconv"
	"strings"

	"gno.land/p/demo/ufmt"
)

const (
	HTML_SKIP_HTML             = 1 << iota // skip preformatted HTML blocks
	HTML_SKIP_STYLE                        // skip embedded <style> elements
	HTML_SKIP_IMAGES                       // skip embedded images
	HTML_SKIP_LINKS                        // skip all links
	HTML_SAFELINK                          // only link to trusted protocols
	HTML_NOFOLLOW_LINKS                    // only link with rel="nofollow"
	HTML_NOREFERRER_LINKS                  // only link with rel="noreferrer"
	HTML_NOOPENER_LINKS                    // only link with rel="noopener"
	HTML_HREF_TARGET_BLANK                 // add a blank target
	HTML_TOC                               // generate a table of contents
	HTML_OMIT_CONTENTS                     // skip the main contents (for a standalone table of contents)
	HTML_COMPLETE_PAGE                     // generate a complete HTML page
	HTML_USE_XHTML                         // generate XHTML output instead of HTML
	HTML_FOOTNOTE_RETURN_LINKS             // generate a link at the end of a footnote to return to the source
)

var (
	alignments = []string{
		"left",
		"right",
		"center",
	}

	htmlEntity = regexp.MustCompile(`&[a-z]{2,5};`)
)

type HtmlRendererParameters struct {
	AbsolutePrefix             string
	FootnoteAnchorPrefix       string
	FootnoteReturnLinkContents string
	HeaderIDPrefix             string
	HeaderIDSuffix             string
}

type Html struct {
	flags    int    // HTML_* options
	closeTag string // how to end singleton tags: either " />" or ">"
	title    string // document title
	css      string // optional css file url (used with HTML_COMPLETE_PAGE)

	parameters HtmlRendererParameters

	tocMarker    int
	headerCount  int
	currentLevel int
	toc          *bytes.Buffer

	headerIDs map[string]int
}

const (
	xhtmlClose = " />"
	htmlClose  = ">"
)

func HtmlRenderer(flags int, title string, css string) Renderer {
	return HtmlRendererWithParameters(flags, title, css, HtmlRendererParameters{})
}

func HtmlRendererWithParameters(flags int, title string,
	css string, renderParameters HtmlRendererParameters,
) Renderer {
	closeTag := htmlClose
	if flags&HTML_USE_XHTML != 0 {
		closeTag = xhtmlClose
	}

	if renderParameters.FootnoteReturnLinkContents == "" {
		renderParameters.FootnoteReturnLinkContents = `<sup>[return]</sup>`
	}

	return &Html{
		flags:      flags,
		closeTag:   closeTag,
		title:      title,
		css:        css,
		parameters: renderParameters,

		headerCount:  0,
		currentLevel: 0,
		toc:          new(bytes.Buffer),

		headerIDs: make(map[string]int),
	}
}

func escapeSingleChar(char byte) (string, bool) {
	if char == '"' {
		return "&quot;", true
	}
	if char == '&' {
		return "&amp;", true
	}
	if char == '<' {
		return "&lt;", true
	}
	if char == '>' {
		return "&gt;", true
	}
	return "", false
}

func attrEscape(out *bytes.Buffer, src []byte) {
	org := 0
	for i, ch := range src {
		if entity, ok := escapeSingleChar(ch); ok {
			if i > org {
				out.Write(src[org:i])
			}
			org = i + 1
			out.WriteString(entity)
		}
	}
	if org < len(src) {
		out.Write(src[org:])
	}
}

func entityEscapeWithSkip(out *bytes.Buffer, src []byte, skipRanges [][]int) {
	end := 0
	for _, rang := range skipRanges {
		attrEscape(out, src[end:rang[0]])
		out.Write(src[rang[0]:rang[1]])
		end = rang[1]
	}
	attrEscape(out, src[end:])
}

func (options *Html) GetFlags() int {
	return options.flags
}

func (options *Html) TitleBlock(out *bytes.Buffer, text []byte) {
	text = bytes.TrimPrefix(text, []byte("% "))
	text = bytes.Replace(text, []byte("\n% "), []byte("\n"), -1)
	out.WriteString("<h1 class=\"title\">")
	out.Write(text)
	out.WriteString("\n</h1>")
}

func (options *Html) Header(out *bytes.Buffer, text func() bool, level int, id string) {
	marker := out.Len()
	doubleSpace(out)

	if id == "" && options.flags&HTML_TOC != 0 {
		id = ufmt.Sprintf("toc_%d", options.headerCount)
	}

	if id != "" {
		id = options.ensureUniqueHeaderID(id)

		if options.parameters.HeaderIDPrefix != "" {
			id = options.parameters.HeaderIDPrefix + id
		}

		if options.parameters.HeaderIDSuffix != "" {
			id = id + options.parameters.HeaderIDSuffix
		}

		out.WriteString(ufmt.Sprintf("<h%d id=\"%s\">", level, id))
	} else {
		out.WriteString(ufmt.Sprintf("<h%d>", level))
	}

	tocMarker := out.Len()
	if !text() {
		out.Truncate(marker)
		return
	}

	if options.flags&HTML_TOC != 0 {
		options.TocHeaderWithAnchor(out.Bytes()[tocMarker:], level, id)
	}

	out.WriteString(ufmt.Sprintf("</h%d>\n", level))
}

func (options *Html) BlockHtml(out *bytes.Buffer, text []byte) {
	if options.flags&HTML_SKIP_HTML != 0 {
		return
	}

	doubleSpace(out)
	out.Write(text)
	out.WriteByte('\n')
}

func (options *Html) HRule(out *bytes.Buffer) {
	doubleSpace(out)
	out.WriteString("<hr")
	out.WriteString(options.closeTag)
	out.WriteByte('\n')
}

func (options *Html) BlockCode(out *bytes.Buffer, text []byte, info string) {
	doubleSpace(out)

	endOfLang := strings.IndexAny(info, "\t ")
	if endOfLang < 0 {
		endOfLang = len(info)
	}
	lang := info[:endOfLang]
	if len(lang) == 0 || lang == "." {
		out.WriteString("<pre><code>")
	} else {
		out.WriteString("<pre><code class=\"language-")
		attrEscape(out, []byte(lang))
		out.WriteString("\">")
	}
	attrEscape(out, text)
	out.WriteString("</code></pre>\n")
}

func (options *Html) BlockQuote(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.WriteString("<blockquote>\n")
	out.Write(text)
	out.WriteString("</blockquote>\n")
}

func (options *Html) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
	doubleSpace(out)
	out.WriteString("<table>\n<thead>\n")
	out.Write(header)
	out.WriteString("</thead>\n\n<tbody>\n")
	out.Write(body)
	out.WriteString("</tbody>\n</table>\n")
}

func (options *Html) TableRow(out *bytes.Buffer, text []byte) {
	doubleSpace(out)
	out.WriteString("<tr>\n")
	out.Write(text)
	out.WriteString("\n</tr>\n")
}

func (options *Html) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	doubleSpace(out)
	switch align {
	case TABLE_ALIGNMENT_LEFT:
		out.WriteString("<th align=\"left\">")
	case TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<th align=\"right\">")
	case TABLE_ALIGNMENT_CENTER:
		out.WriteString("<th align=\"center\">")
	default:
		out.WriteString("<th>")
	}

	out.Write(text)
	out.WriteString("</th>")
}

func (options *Html) TableCell(out *bytes.Buffer, text []byte, align int) {
	doubleSpace(out)
	switch align {
	case TABLE_ALIGNMENT_LEFT:
		out.WriteString("<td align=\"left\">")
	case TABLE_ALIGNMENT_RIGHT:
		out.WriteString("<td align=\"right\">")
	case TABLE_ALIGNMENT_CENTER:
		out.WriteString("<td align=\"center\">")
	default:
		out.WriteString("<td>")
	}

	out.Write(text)
	out.WriteString("</td>")
}

func (options *Html) Footnotes(out *bytes.Buffer, text func() bool) {
	out.WriteString("<div class=\"footnotes\">\n")
	options.HRule(out)
	options.List(out, text, LIST_TYPE_ORDERED)
	out.WriteString("</div>\n")
}

func (options *Html) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
	if flags&LIST_ITEM_CONTAINS_BLOCK != 0 || flags&LIST_ITEM_BEGINNING_OF_LIST != 0 {
		doubleSpace(out)
	}
	slug := slugify(name)
	out.WriteString(`<li id="`)
	out.WriteString(`fn:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`">`)
	out.Write(text)
	if options.flags&HTML_FOOTNOTE_RETURN_LINKS != 0 {
		out.WriteString(` <a class="footnote-return" href="#`)
		out.WriteString(`fnref:`)
		out.WriteString(options.parameters.FootnoteAnchorPrefix)
		out.Write(slug)
		out.WriteString(`">`)
		out.WriteString(options.parameters.FootnoteReturnLinkContents)
		out.WriteString(`</a>`)
	}
	out.WriteString("</li>\n")
}

func (options *Html) List(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	doubleSpace(out)

	if flags&LIST_TYPE_DEFINITION != 0 {
		out.WriteString("<dl>")
	} else if flags&LIST_TYPE_ORDERED != 0 {
		out.WriteString("<ol>")
	} else {
		out.WriteString("<ul>")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if flags&LIST_TYPE_DEFINITION != 0 {
		out.WriteString("</dl>\n")
	} else if flags&LIST_TYPE_ORDERED != 0 {
		out.WriteString("</ol>\n")
	} else {
		out.WriteString("</ul>\n")
	}
}

func (options *Html) ListItem(out *bytes.Buffer, text []byte, flags int) {
	if (flags&LIST_ITEM_CONTAINS_BLOCK != 0 && flags&LIST_TYPE_DEFINITION == 0) ||
		flags&LIST_ITEM_BEGINNING_OF_LIST != 0 {
		doubleSpace(out)
	}
	if flags&LIST_TYPE_TERM != 0 {
		out.WriteString("<dt>")
	} else if flags&LIST_TYPE_DEFINITION != 0 {
		out.WriteString("<dd>")
	} else {
		out.WriteString("<li>")
	}
	out.Write(text)
	if flags&LIST_TYPE_TERM != 0 {
		out.WriteString("</dt>\n")
	} else if flags&LIST_TYPE_DEFINITION != 0 {
		out.WriteString("</dd>\n")
	} else {
		out.WriteString("</li>\n")
	}
}

func (options *Html) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()
	doubleSpace(out)

	out.WriteString("<p>")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("</p>\n")
}

func (options *Html) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	skipRanges := htmlEntity.FindAllIndex(link, -1)
	if options.flags&HTML_SAFELINK != 0 && !isSafeLink(link) && kind != LINK_TYPE_EMAIL {
		out.WriteString("<tt>")
		entityEscapeWithSkip(out, link, skipRanges)
		out.WriteString("</tt>")
		return
	}

	out.WriteString("<a href=\"")
	if kind == LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	} else {
		options.maybeWriteAbsolutePrefix(out, link)
	}

	entityEscapeWithSkip(out, link, skipRanges)

	var relAttrs []string
	if options.flags&HTML_NOFOLLOW_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "nofollow")
	}
	if options.flags&HTML_NOREFERRER_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "noreferrer")
	}
	if options.flags&HTML_NOOPENER_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "noopener")
	}
	if len(relAttrs) > 0 {
		out.WriteString(ufmt.Sprintf("\" rel=\"%s", strings.Join(relAttrs, " ")))
	}

	if options.flags&HTML_HREF_TARGET_BLANK != 0 && !isRelativeLink(link) {
		out.WriteString("\" target=\"_blank")
	}

	out.WriteString("\">")

	switch {
	case bytes.HasPrefix(link, []byte("mailto://")):
		attrEscape(out, link[len("mailto://"):])
	case bytes.HasPrefix(link, []byte("mailto:")):
		attrEscape(out, link[len("mailto:"):])
	default:
		entityEscapeWithSkip(out, link, skipRanges)
	}

	out.WriteString("</a>")
}

func (options *Html) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("<code>")
	attrEscape(out, text)
	out.WriteString("</code>")
}

func (options *Html) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong>")
	out.Write(text)
	out.WriteString("</strong>")
}

func (options *Html) Emphasis(out *bytes.Buffer, text []byte) {
	if len(text) == 0 {
		return
	}
	out.WriteString("<em>")
	out.Write(text)
	out.WriteString("</em>")
}

func (options *Html) maybeWriteAbsolutePrefix(out *bytes.Buffer, link []byte) {
	if options.parameters.AbsolutePrefix != "" && isRelativeLink(link) && link[0] != '.' {
		out.WriteString(options.parameters.AbsolutePrefix)
		if link[0] != '/' {
			out.WriteByte('/')
		}
	}
}

func (options *Html) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
	if options.flags&HTML_SKIP_IMAGES != 0 {
		return
	}

	out.WriteString("<img src=\"")
	options.maybeWriteAbsolutePrefix(out, link)
	attrEscape(out, link)
	out.WriteString("\" alt=\"")
	if len(alt) > 0 {
		attrEscape(out, alt)
	}
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		attrEscape(out, title)
	}

	out.WriteByte('"')
	out.WriteString(options.closeTag)
}

func (options *Html) LineBreak(out *bytes.Buffer) {
	out.WriteString("<br")
	out.WriteString(options.closeTag)
	out.WriteByte('\n')
}

func (options *Html) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	if options.flags&HTML_SKIP_LINKS != 0 {
		out.WriteString("<tt>")
		attrEscape(out, content)
		out.WriteString("</tt>")
		return
	}

	if options.flags&HTML_SAFELINK != 0 && !isSafeLink(link) {
		out.WriteString("<tt>")
		attrEscape(out, content)
		out.WriteString("</tt>")
		return
	}

	out.WriteString("<a href=\"")
	options.maybeWriteAbsolutePrefix(out, link)
	attrEscape(out, link)
	if len(title) > 0 {
		out.WriteString("\" title=\"")
		attrEscape(out, title)
	}
	var relAttrs []string
	if options.flags&HTML_NOFOLLOW_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "nofollow")
	}
	if options.flags&HTML_NOREFERRER_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "noreferrer")
	}
	if options.flags&HTML_NOOPENER_LINKS != 0 && !isRelativeLink(link) {
		relAttrs = append(relAttrs, "noopener")
	}
	if len(relAttrs) > 0 {
		out.WriteString(ufmt.Sprintf("\" rel=\"%s", strings.Join(relAttrs, " ")))
	}

	if options.flags&HTML_HREF_TARGET_BLANK != 0 && !isRelativeLink(link) {
		out.WriteString("\" target=\"_blank")
	}

	out.WriteString("\">")
	out.Write(content)
	out.WriteString("</a>")
	return
}

func (options *Html) RawHtmlTag(out *bytes.Buffer, text []byte) {
	if options.flags&HTML_SKIP_HTML != 0 {
		return
	}
	if options.flags&HTML_SKIP_STYLE != 0 && isHtmlTag(text, "style") {
		return
	}
	if options.flags&HTML_SKIP_LINKS != 0 && isHtmlTag(text, "a") {
		return
	}
	if options.flags&HTML_SKIP_IMAGES != 0 && isHtmlTag(text, "img") {
		return
	}
	out.Write(text)
}

func (options *Html) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("<strong><em>")
	out.Write(text)
	out.WriteString("</em></strong>")
}

func (options *Html) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("<del>")
	out.Write(text)
	out.WriteString("</del>")
}

func (options *Html) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
	slug := slugify(ref)
	out.WriteString(`<sup class="footnote-ref" id="`)
	out.WriteString(`fnref:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`"><a href="#`)
	out.WriteString(`fn:`)
	out.WriteString(options.parameters.FootnoteAnchorPrefix)
	out.Write(slug)
	out.WriteString(`">`)
	out.WriteString(strconv.Itoa(id))
	out.WriteString(`</a></sup>`)
}

func (options *Html) Entity(out *bytes.Buffer, entity []byte) {
	out.Write(entity)
}

func (options *Html) NormalText(out *bytes.Buffer, text []byte) {
	attrEscape(out, text)
}

func (options *Html) DocumentHeader(out *bytes.Buffer) {
	if options.flags&HTML_COMPLETE_PAGE == 0 {
		return
	}

	ending := ""
	if options.flags&HTML_USE_XHTML != 0 {
		out.WriteString("<!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" ")
		out.WriteString("\"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\n")
		out.WriteString("<html xmlns=\"http://www.w3.org/1999/xhtml\">\n")
		ending = " /"
	} else {
		out.WriteString("<!DOCTYPE html>\n")
		out.WriteString("<html>\n")
	}
	out.WriteString("<head>\n")
	out.WriteString("  <title>")
	options.NormalText(out, []byte(options.title))
	out.WriteString("</title>\n")
	out.WriteString("  <meta name=\"GENERATOR\" content=\"Blackfriday Markdown Processor v")
	out.WriteString(VERSION)
	out.WriteString("\"")
	out.WriteString(ending)
	out.WriteString(">\n")
	out.WriteString("  <meta charset=\"utf-8\"")
	out.WriteString(ending)
	out.WriteString(">\n")
	if options.css != "" {
		out.WriteString("  <link rel=\"stylesheet\" type=\"text/css\" href=\"")
		attrEscape(out, []byte(options.css))
		out.WriteString("\"")
		out.WriteString(ending)
		out.WriteString(">\n")
	}
	out.WriteString("</head>\n")
	out.WriteString("<body>\n")

	options.tocMarker = out.Len()
}

func (options *Html) DocumentFooter(out *bytes.Buffer) {
	if options.flags&HTML_TOC != 0 {
		options.TocFinalize()

		var temp bytes.Buffer

		temp.Write(out.Bytes()[options.tocMarker:])

		out.Truncate(options.tocMarker)

		if options.flags&HTML_COMPLETE_PAGE != 0 {
			out.WriteByte('\n')
		}

		out.WriteString("<nav>\n")
		out.Write(options.toc.Bytes())
		out.WriteString("</nav>\n")

		if options.flags&HTML_COMPLETE_PAGE == 0 && options.flags&HTML_OMIT_CONTENTS == 0 {
			out.WriteByte('\n')
		}

		if options.flags&HTML_OMIT_CONTENTS == 0 {
			out.Write(temp.Bytes())
		}
	}

	if options.flags&HTML_COMPLETE_PAGE != 0 {
		out.WriteString("\n</body>\n")
		out.WriteString("</html>\n")
	}
}

func (options *Html) TocHeaderWithAnchor(text []byte, level int, anchor string) {
	for level > options.currentLevel {
		switch {
		case bytes.HasSuffix(options.toc.Bytes(), []byte("</li>\n")):
			size := options.toc.Len()
			options.toc.Truncate(size - len("</li>\n"))

		case options.currentLevel > 0:
			options.toc.WriteString("<li>")
		}
		if options.toc.Len() > 0 {
			options.toc.WriteByte('\n')
		}
		options.toc.WriteString("<ul>\n")
		options.currentLevel++
	}

	for level < options.currentLevel {
		options.toc.WriteString("</ul>")
		if options.currentLevel > 1 {
			options.toc.WriteString("</li>\n")
		}
		options.currentLevel--
	}

	options.toc.WriteString("<li><a href=\"#")
	if anchor != "" {
		options.toc.WriteString(anchor)
	} else {
		options.toc.WriteString("toc_")
		options.toc.WriteString(strconv.Itoa(options.headerCount))
	}
	options.toc.WriteString("\">")
	options.headerCount++

	options.toc.Write(text)

	options.toc.WriteString("</a></li>\n")
}

func (options *Html) TocHeader(text []byte, level int) {
	options.TocHeaderWithAnchor(text, level, "")
}

func (options *Html) TocFinalize() {
	for options.currentLevel > 1 {
		options.toc.WriteString("</ul></li>\n")
		options.currentLevel--
	}

	if options.currentLevel > 0 {
		options.toc.WriteString("</ul>\n")
	}
}

func isHtmlTag(tag []byte, tagname string) bool {
	found, _ := findHtmlTagPos(tag, tagname)
	return found
}

func skipUntilCharIgnoreQuotes(html []byte, start int, char byte) int {
	inSingleQuote := false
	inDoubleQuote := false
	inGraveQuote := false
	i := start
	for i < len(html) {
		switch {
		case html[i] == char && !inSingleQuote && !inDoubleQuote && !inGraveQuote:
			return i
		case html[i] == '\'':
			inSingleQuote = !inSingleQuote
		case html[i] == '"':
			inDoubleQuote = !inDoubleQuote
		case html[i] == '`':
			inGraveQuote = !inGraveQuote
		}
		i++
	}
	return start
}

func findHtmlTagPos(tag []byte, tagname string) (bool, int) {
	i := 0
	if i < len(tag) && tag[0] != '<' {
		return false, -1
	}
	i++
	i = skipSpace(tag, i)

	if i < len(tag) && tag[i] == '/' {
		i++
	}

	i = skipSpace(tag, i)
	j := 0
	for ; i < len(tag); i, j = i+1, j+1 {
		if j >= len(tagname) {
			break
		}

		if strings.ToLower(string(tag[i]))[0] != tagname[j] {
			return false, -1
		}
	}

	if i == len(tag) {
		return false, -1
	}

	rightAngle := skipUntilCharIgnoreQuotes(tag, i, '>')
	if rightAngle > i {
		return true, rightAngle
	}

	return false, -1
}

func skipUntilChar(text []byte, start int, char byte) int {
	i := start
	for i < len(text) && text[i] != char {
		i++
	}
	return i
}

func skipSpace(tag []byte, i int) int {
	for i < len(tag) && isspace(tag[i]) {
		i++
	}
	return i
}

func skipChar(data []byte, start int, char byte) int {
	i := start
	for i < len(data) && data[i] == char {
		i++
	}
	return i
}

func doubleSpace(out *bytes.Buffer) {
	if out.Len() > 0 {
		out.WriteByte('\n')
	}
}

func isRelativeLink(link []byte) (yes bool) {
	if link[0] == '#' {
		return true
	}

	if len(link) >= 2 && link[0] == '/' && link[1] != '/' {
		return true
	}

	if len(link) == 1 && link[0] == '/' {
		return true
	}

	if bytes.HasPrefix(link, []byte("./")) {
		return true
	}

	if bytes.HasPrefix(link, []byte("../")) {
		return true
	}

	return false
}

func (options *Html) ensureUniqueHeaderID(id string) string {
	for count, found := options.headerIDs[id]; found; count, found = options.headerIDs[id] {
		tmp := ufmt.Sprintf("%s-%d", id, count+1)

		if _, tmpFound := options.headerIDs[tmp]; !tmpFound {
			options.headerIDs[id] = count + 1
			id = tmp
		} else {
			id = id + "-1"
		}
	}

	if _, found := options.headerIDs[id]; !found {
		options.headerIDs[id] = 0
	}

	return id
}
