package varmeta_md

import (
	"bytes"
	"strings"
	"unicode"
)

func (p *parser) block(out *bytes.Buffer, data []byte) {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		panic("block input is missing terminating newline")
	}

	if p.nesting >= p.maxNesting {
		return
	}
	p.nesting++

	for len(data) > 0 {
		if p.isPrefixHeader(data) {
			data = data[p.prefixHeader(out, data):]
			continue
		}

		if data[0] == '<' {
			if i := p.html(out, data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		if p.flags&EXTENSION_TITLEBLOCK != 0 {
			if data[0] == '%' {
				if i := p.titleBlock(out, data, true); i > 0 {
					data = data[i:]
					continue
				}
			}
		}

		if i := p.isEmpty(data); i > 0 {
			data = data[i:]
			continue
		}

		if p.codePrefix(data) > 0 {
			data = data[p.code(out, data):]
			continue
		}

		if p.flags&EXTENSION_FENCED_CODE != 0 {
			if i := p.fencedCodeBlock(out, data, true); i > 0 {
				data = data[i:]
				continue
			}
		}

		if p.isHRule(data) {
			p.r.HRule(out)
			var i int
			for i = 0; data[i] != '\n'; i++ {
			}
			data = data[i:]
			continue
		}

		if p.quotePrefix(data) > 0 {
			data = data[p.quote(out, data):]
			continue
		}

		if p.flags&EXTENSION_TABLES != 0 {
			if i := p.table(out, data); i > 0 {
				data = data[i:]
				continue
			}
		}

		if p.uliPrefix(data) > 0 {
			data = data[p.list(out, data, 0):]
			continue
		}

		if p.oliPrefix(data) > 0 {
			data = data[p.list(out, data, LIST_TYPE_ORDERED):]
			continue
		}

		if p.flags&EXTENSION_DEFINITION_LISTS != 0 {
			if p.dliPrefix(data) > 0 {
				data = data[p.list(out, data, LIST_TYPE_DEFINITION):]
				continue
			}
		}

		data = data[p.paragraph(out, data):]
	}

	p.nesting--
}

func (p *parser) isPrefixHeader(data []byte) bool {
	if data[0] != '#' {
		return false
	}

	if p.flags&EXTENSION_SPACE_HEADERS != 0 {
		level := 0
		for level < 6 && data[level] == '#' {
			level++
		}
		if data[level] != ' ' {
			return false
		}
	}
	return true
}

func (p *parser) prefixHeader(out *bytes.Buffer, data []byte) int {
	level := 0
	for level < 6 && data[level] == '#' {
		level++
	}
	i := skipChar(data, level, ' ')
	end := skipUntilChar(data, i, '\n')
	skip := end
	id := ""
	if p.flags&EXTENSION_HEADER_IDS != 0 {
		j, k := 0, 0
		for j = i; j < end-1 && (data[j] != '{' || data[j+1] != '#'); j++ {
		}
		for k = j + 1; k < end && data[k] != '}'; k++ {
		}
		if j < end && k < end {
			id = string(data[j+2 : k])
			end = j
			skip = k + 1
			for end > 0 && data[end-1] == ' ' {
				end--
			}
		}
	}
	for end > 0 && data[end-1] == '#' {
		if isBackslashEscaped(data, end-1) {
			break
		}
		end--
	}
	for end > 0 && data[end-1] == ' ' {
		end--
	}
	if end > i {
		if id == "" && p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
			id = SanitizedAnchorName(string(data[i:end]))
		}
		work := func() bool {
			p.inline(out, data[i:end])
			return true
		}
		p.r.Header(out, work, level, id)
	}
	return skip
}

func (p *parser) isUnderlinedHeader(data []byte) int {
	if data[0] == '=' {
		i := skipChar(data, 1, '=')
		i = skipChar(data, i, ' ')
		if data[i] == '\n' {
			return 1
		} else {
			return 0
		}
	}

	if data[0] == '-' {
		i := skipChar(data, 1, '-')
		i = skipChar(data, i, ' ')
		if data[i] == '\n' {
			return 2
		} else {
			return 0
		}
	}

	return 0
}

func (p *parser) titleBlock(out *bytes.Buffer, data []byte, doRender bool) int {
	if data[0] != '%' {
		return 0
	}
	splitData := bytes.Split(data, []byte("\n"))
	var i int
	for idx, b := range splitData {
		if !bytes.HasPrefix(b, []byte("%")) {
			i = idx
			break
		}
	}

	data = bytes.Join(splitData[0:i], []byte("\n"))
	p.r.TitleBlock(out, data)

	return len(data)
}

func (p *parser) html(out *bytes.Buffer, data []byte, doRender bool) int {
	var i, j int

	if data[0] != '<' {
		return 0
	}
	curtag, tagfound := p.htmlFindTag(data[1:])

	if !tagfound {
		if size := p.htmlComment(out, data, doRender); size > 0 {
			return size
		}

		if size := p.htmlHr(out, data, doRender); size > 0 {
			return size
		}

		if size := p.htmlCDATA(out, data, doRender); size > 0 {
			return size
		}

		return 0
	}

	found := false

	if !found && curtag != "ins" && curtag != "del" {
		i = 1
		for i < len(data) {
			i++
			for i < len(data) && !(data[i-1] == '<' && data[i] == '/') {
				i++
			}

			if i+2+len(curtag) >= len(data) {
				break
			}

			j = p.htmlFindEnd(curtag, data[i-1:])

			if j > 0 {
				i += j - 1
				found = true
				break
			}
		}
	}

	if !found {
		return 0
	}

	if doRender {
		end := i
		for end > 0 && data[end-1] == '\n' {
			end--
		}
		p.r.BlockHtml(out, data[:end])
	}

	return i
}

func (p *parser) renderHTMLBlock(out *bytes.Buffer, data []byte, start int, doRender bool) int {
	if i := p.isEmpty(data[start:]); i > 0 {
		size := start + i
		if doRender {
			end := size
			for end > 0 && data[end-1] == '\n' {
				end--
			}
			p.r.BlockHtml(out, data[:end])
		}
		return size
	}
	return 0
}

func (p *parser) htmlComment(out *bytes.Buffer, data []byte, doRender bool) int {
	i := p.inlineHTMLComment(out, data)
	return p.renderHTMLBlock(out, data, i, doRender)
}

func (p *parser) htmlCDATA(out *bytes.Buffer, data []byte, doRender bool) int {
	const cdataTag = "<![cdata["
	const cdataTagLen = len(cdataTag)
	if len(data) < cdataTagLen+1 {
		return 0
	}
	if !bytes.Equal(bytes.ToLower(data[:cdataTagLen]), []byte(cdataTag)) {
		return 0
	}
	i := cdataTagLen
	for i < len(data) && !(data[i-2] == ']' && data[i-1] == ']' && data[i] == '>') {
		i++
	}
	i++
	if i >= len(data) {
		return 0
	}
	return p.renderHTMLBlock(out, data, i, doRender)
}

func (p *parser) htmlHr(out *bytes.Buffer, data []byte, doRender bool) int {
	if data[0] != '<' || (data[1] != 'h' && data[1] != 'H') || (data[2] != 'r' && data[2] != 'R') {
		return 0
	}
	if data[3] != ' ' && data[3] != '/' && data[3] != '>' {
		return 0
	}

	i := 3
	for data[i] != '>' && data[i] != '\n' {
		i++
	}

	if data[i] == '>' {
		return p.renderHTMLBlock(out, data, i+1, doRender)
	}

	return 0
}

func (p *parser) htmlFindTag(data []byte) (string, bool) {
	i := 0
	for isalnum(data[i]) {
		i++
	}
	key := string(data[:i])
	if _, ok := blockTags[key]; ok {
		return key, true
	}
	return "", false
}

func (p *parser) htmlFindEnd(tag string, data []byte) int {
	closetag := []byte("</" + tag + ">")
	if !bytes.HasPrefix(data, closetag) {
		return 0
	}
	i := len(closetag)

	skip := 0
	if skip = p.isEmpty(data[i:]); skip == 0 {
		return 0
	}
	i += skip
	skip = 0

	if i >= len(data) {
		return i
	}

	if p.flags&EXTENSION_LAX_HTML_BLOCKS != 0 {
		return i
	}
	if skip = p.isEmpty(data[i:]); skip == 0 {
		return 0
	}

	return i + skip
}

func (*parser) isEmpty(data []byte) int {
	if len(data) == 0 {
		return 0
	}

	var i int
	for i = 0; i < len(data) && data[i] != '\n'; i++ {
		if data[i] != ' ' && data[i] != '\t' {
			return 0
		}
	}
	return i + 1
}

func (*parser) isHRule(data []byte) bool {
	i := 0

	for i < 3 && data[i] == ' ' {
		i++
	}

	if data[i] != '*' && data[i] != '-' && data[i] != '_' {
		return false
	}
	c := data[i]

	n := 0
	for data[i] != '\n' {
		switch {
		case data[i] == c:
			n++
		case data[i] != ' ':
			return false
		}
		i++
	}

	return n >= 3
}

func isFenceLine(data []byte, info *string, oldmarker string, newlineOptional bool) (end int, marker string) {
	i, size := 0, 0

	for i < len(data) && i < 3 && data[i] == ' ' {
		i++
	}

	if i >= len(data) {
		return 0, ""
	}
	if data[i] != '~' && data[i] != '`' {
		return 0, ""
	}

	c := data[i]

	for i < len(data) && data[i] == c {
		size++
		i++
	}

	if size < 3 {
		return 0, ""
	}
	marker = string(data[i-size : i])

	if oldmarker != "" && marker != oldmarker {
		return 0, ""
	}

	if info != nil {
		infoLength := 0
		i = skipChar(data, i, ' ')

		if i >= len(data) {
			if newlineOptional && i == len(data) {
				return i, marker
			}
			return 0, ""
		}

		infoStart := i

		if data[i] == '{' {
			i++
			infoStart++

			for i < len(data) && data[i] != '}' && data[i] != '\n' {
				infoLength++
				i++
			}

			if i >= len(data) || data[i] != '}' {
				return 0, ""
			}

			for infoLength > 0 && isspace(data[infoStart]) {
				infoStart++
				infoLength--
			}

			for infoLength > 0 && isspace(data[infoStart+infoLength-1]) {
				infoLength--
			}

			i++
		} else {
			for i < len(data) && !isverticalspace(data[i]) {
				infoLength++
				i++
			}
		}

		*info = strings.TrimSpace(string(data[infoStart : infoStart+infoLength]))
	}

	i = skipChar(data, i, ' ')
	if i >= len(data) {
		if newlineOptional {
			return i, marker
		}
		return 0, ""
	}
	if data[i] == '\n' {
		i++
	}

	return i, marker
}

func (p *parser) fencedCodeBlock(out *bytes.Buffer, data []byte, doRender bool) int {
	var infoString string
	beg, marker := isFenceLine(data, &infoString, "", false)
	if beg == 0 || beg >= len(data) {
		return 0
	}

	var work bytes.Buffer

	for {
		newlineOptional := !doRender
		fenceEnd, _ := isFenceLine(data[beg:], nil, marker, newlineOptional)
		if fenceEnd != 0 {
			beg += fenceEnd
			break
		}

		end := skipUntilChar(data, beg, '\n') + 1

		if end >= len(data) {
			return 0
		}

		if doRender {
			work.Write(data[beg:end])
		}
		beg = end
	}

	if doRender {
		p.r.BlockCode(out, work.Bytes(), infoString)
	}

	return beg
}

func (p *parser) table(out *bytes.Buffer, data []byte) int {
	var header bytes.Buffer
	i, columns := p.tableHeader(&header, data)
	if i == 0 {
		return 0
	}

	var body bytes.Buffer

	for i < len(data) {
		pipes, rowStart := 0, i
		for ; data[i] != '\n'; i++ {
			if data[i] == '|' {
				pipes++
			}
		}

		if pipes == 0 {
			i = rowStart
			break
		}

		i++
		p.tableRow(&body, data[rowStart:i], columns, false)
	}

	p.r.Table(out, header.Bytes(), body.Bytes(), columns)

	return i
}

func isBackslashEscaped(data []byte, i int) bool {
	backslashes := 0
	for i-backslashes-1 >= 0 && data[i-backslashes-1] == '\\' {
		backslashes++
	}
	return backslashes&1 == 1
}

func (p *parser) tableHeader(out *bytes.Buffer, data []byte) (size int, columns []int) {
	i := 0
	colCount := 1
	for i = 0; data[i] != '\n'; i++ {
		if data[i] == '|' && !isBackslashEscaped(data, i) {
			colCount++
		}
	}

	if colCount == 1 {
		return
	}

	header := data[:i+1]

	if data[0] == '|' {
		colCount--
	}
	if i > 2 && data[i-1] == '|' && !isBackslashEscaped(data, i-1) {
		colCount--
	}

	columns = make([]int, colCount)

	i++
	if i >= len(data) {
		return
	}

	if data[i] == '|' && !isBackslashEscaped(data, i) {
		i++
	}
	i = skipChar(data, i, ' ')

	col := 0
	for data[i] != '\n' {
		dashes := 0

		if data[i] == ':' {
			i++
			columns[col] |= TABLE_ALIGNMENT_LEFT
			dashes++
		}
		for data[i] == '-' {
			i++
			dashes++
		}
		if data[i] == ':' {
			i++
			columns[col] |= TABLE_ALIGNMENT_RIGHT
			dashes++
		}
		for data[i] == ' ' {
			i++
		}

		switch {
		case dashes < 3:
			return

		case data[i] == '|' && !isBackslashEscaped(data, i):
			col++
			i++
			for data[i] == ' ' {
				i++
			}

			if col >= colCount && data[i] != '\n' {
				return
			}

		case (data[i] != '|' || isBackslashEscaped(data, i)) && col+1 < colCount:
			return

		case data[i] == '\n':
			col++

		default:
			return
		}
	}
	if col != colCount {
		return
	}

	p.tableRow(out, header, columns, true)
	size = i + 1
	return
}

func (p *parser) tableRow(out *bytes.Buffer, data []byte, columns []int, header bool) {
	i, col := 0, 0
	var rowWork bytes.Buffer

	if data[i] == '|' && !isBackslashEscaped(data, i) {
		i++
	}

	for col = 0; col < len(columns) && i < len(data); col++ {
		for data[i] == ' ' {
			i++
		}

		cellStart := i

		for (data[i] != '|' || isBackslashEscaped(data, i)) && data[i] != '\n' {
			i++
		}

		cellEnd := i

		i++

		for cellEnd > cellStart && data[cellEnd-1] == ' ' {
			cellEnd--
		}

		var cellWork bytes.Buffer
		p.inline(&cellWork, data[cellStart:cellEnd])

		if header {
			p.r.TableHeaderCell(&rowWork, cellWork.Bytes(), columns[col])
		} else {
			p.r.TableCell(&rowWork, cellWork.Bytes(), columns[col])
		}
	}

	for ; col < len(columns); col++ {
		if header {
			p.r.TableHeaderCell(&rowWork, nil, columns[col])
		} else {
			p.r.TableCell(&rowWork, nil, columns[col])
		}
	}

	p.r.TableRow(out, rowWork.Bytes())
}

func (p *parser) quotePrefix(data []byte) int {
	i := 0
	for i < 3 && data[i] == ' ' {
		i++
	}
	if data[i] == '>' {
		if data[i+1] == ' ' {
			return i + 2
		}
		return i + 1
	}
	return 0
}

func (p *parser) terminateBlockquote(data []byte, beg, end int) bool {
	if p.isEmpty(data[beg:]) <= 0 {
		return false
	}
	if end >= len(data) {
		return true
	}
	return p.quotePrefix(data[end:]) == 0 && p.isEmpty(data[end:]) == 0
}

func (p *parser) quote(out *bytes.Buffer, data []byte) int {
	var raw bytes.Buffer
	beg, end := 0, 0
	for beg < len(data) {
		end = beg
		for data[end] != '\n' {
			if p.flags&EXTENSION_FENCED_CODE != 0 {
				if i := p.fencedCodeBlock(out, data[end:], false); i > 0 {
					end += i - 1
					break
				}
			}
			end++
		}
		end++

		if pre := p.quotePrefix(data[beg:]); pre > 0 {
			beg += pre
		} else if p.terminateBlockquote(data, beg, end) {
			break
		}

		raw.Write(data[beg:end])
		beg = end
	}

	var cooked bytes.Buffer
	p.block(&cooked, raw.Bytes())
	p.r.BlockQuote(out, cooked.Bytes())
	return end
}

func (p *parser) codePrefix(data []byte) int {
	if data[0] == ' ' && data[1] == ' ' && data[2] == ' ' && data[3] == ' ' {
		return 4
	}
	return 0
}

func (p *parser) code(out *bytes.Buffer, data []byte) int {
	var work bytes.Buffer

	i := 0
	for i < len(data) {
		beg := i
		for data[i] != '\n' {
			i++
		}
		i++

		blankline := p.isEmpty(data[beg:i]) > 0
		if pre := p.codePrefix(data[beg:i]); pre > 0 {
			beg += pre
		} else if !blankline {
			i = beg
			break
		}

		if blankline {
			work.WriteByte('\n')
		} else {
			work.Write(data[beg:i])
		}
	}

	workbytes := work.Bytes()
	eol := len(workbytes)
	for eol > 0 && workbytes[eol-1] == '\n' {
		eol--
	}
	if eol != len(workbytes) {
		work.Truncate(eol)
	}

	work.WriteByte('\n')

	p.r.BlockCode(out, work.Bytes(), "")

	return i
}

func (p *parser) uliPrefix(data []byte) int {
	i := 0

	for i < 3 && data[i] == ' ' {
		i++
	}

	if (data[i] != '*' && data[i] != '+' && data[i] != '-') ||
		data[i+1] != ' ' {
		return 0
	}
	return i + 2
}

func (p *parser) oliPrefix(data []byte) int {
	i := 0

	for i < 3 && data[i] == ' ' {
		i++
	}

	start := i
	for data[i] >= '0' && data[i] <= '9' {
		i++
	}

	if start == i || data[i] != '.' || data[i+1] != ' ' {
		return 0
	}
	return i + 2
}

func (p *parser) dliPrefix(data []byte) int {
	i := 0

	if data[i] != ':' || data[i+1] != ' ' {
		return 0
	}
	for data[i] == ' ' {
		i++
	}
	return i + 2
}

func (p *parser) list(out *bytes.Buffer, data []byte, flags int) int {
	i := 0
	flags |= LIST_ITEM_BEGINNING_OF_LIST
	work := func() bool {
		for i < len(data) {
			skip := p.listItem(out, data[i:], &flags)
			i += skip

			if skip == 0 || flags&LIST_ITEM_END_OF_LIST != 0 {
				break
			}
			if flags&LIST_ITEM_BEGINNING_OF_LIST != 0 {
				flags -= LIST_ITEM_BEGINNING_OF_LIST
			}

		}
		return true
	}

	p.r.List(out, work, flags)
	return i
}

func (p *parser) listItem(out *bytes.Buffer, data []byte, flags *int) int {
	itemIndent := 0
	for itemIndent < 3 && data[itemIndent] == ' ' {
		itemIndent++
	}

	i := p.uliPrefix(data)
	if i == 0 {
		i = p.oliPrefix(data)
	}
	if i == 0 {
		i = p.dliPrefix(data)
		if i > 0 {
			if *flags&LIST_TYPE_TERM != 0 {
				*flags -= LIST_TYPE_TERM
			}
		}
	}
	if i == 0 {
		if *flags&LIST_TYPE_DEFINITION != 0 {
			*flags |= LIST_TYPE_TERM
		} else {
			return 0
		}
	}

	for data[i] == ' ' {
		i++
	}

	line := i
	for i > 0 && data[i-1] != '\n' {
		i++
	}

	containsBlankLine := false
	sublist := 0
	codeBlockMarker := ""
	if p.flags&EXTENSION_FENCED_CODE != 0 && i > line {
		_, codeBlockMarker = isFenceLine(data[line:i], nil, "", false)
	}

	var raw bytes.Buffer

	raw.Write(data[line:i])
	line = i

gatherlines:
	for line < len(data) {
		i++

		for data[i-1] != '\n' {
			i++
		}
		if p.isEmpty(data[line:i]) > 0 {
			containsBlankLine = true
			raw.Write(data[line:i])
			line = i
			continue
		}

		indent := 0
		for indent < 4 && line+indent < i && data[line+indent] == ' ' {
			indent++
		}

		chunk := data[line+indent : i]

		if p.flags&EXTENSION_FENCED_CODE != 0 {
			_, marker := isFenceLine(chunk, nil, codeBlockMarker, false)
			if marker != "" {
				if codeBlockMarker == "" {
					codeBlockMarker = marker
				} else {
					*flags |= LIST_ITEM_CONTAINS_BLOCK
					codeBlockMarker = ""
				}
			}
			if codeBlockMarker != "" || marker != "" {
				raw.Write(data[line+indent : i])
				line = i
				continue gatherlines
			}
		}

		switch {
		case (p.uliPrefix(chunk) > 0 && !p.isHRule(chunk)) ||
			p.oliPrefix(chunk) > 0 ||
			p.dliPrefix(chunk) > 0:

			if containsBlankLine {
				if indent <= itemIndent &&
					((*flags&LIST_TYPE_ORDERED != 0 && p.uliPrefix(chunk) > 0) ||
						(*flags&LIST_TYPE_ORDERED == 0 && p.oliPrefix(chunk) > 0)) {

					*flags |= LIST_ITEM_END_OF_LIST
					break gatherlines
				}
				*flags |= LIST_ITEM_CONTAINS_BLOCK
			}

			if indent <= itemIndent {
				break gatherlines
			}

			if sublist == 0 {
				sublist = raw.Len()
			}

		case p.isPrefixHeader(chunk):
			if containsBlankLine && indent < 4 {
				*flags |= LIST_ITEM_END_OF_LIST
				break gatherlines
			}
			*flags |= LIST_ITEM_CONTAINS_BLOCK

		case containsBlankLine && indent < 4:
			if *flags&LIST_TYPE_DEFINITION != 0 && i < len(data)-1 {
				next := i
				for data[next] != '\n' {
					next++
				}
				for next < len(data)-1 && data[next] == '\n' {
					next++
				}
				if i < len(data)-1 && data[i] != ':' && data[next] != ':' {
					*flags |= LIST_ITEM_END_OF_LIST
				}
			} else {
				*flags |= LIST_ITEM_END_OF_LIST
			}
			break gatherlines

		case containsBlankLine:
			*flags |= LIST_ITEM_CONTAINS_BLOCK
		}

		containsBlankLine = false

		raw.Write(data[line+indent : i])

		line = i
	}

	if line >= len(data) {
		*flags |= LIST_ITEM_END_OF_LIST
	}

	rawBytes := raw.Bytes()

	var cooked bytes.Buffer
	if *flags&LIST_ITEM_CONTAINS_BLOCK != 0 && *flags&LIST_TYPE_TERM == 0 {
		if sublist > 0 {
			p.block(&cooked, rawBytes[:sublist])
			p.block(&cooked, rawBytes[sublist:])
		} else {
			p.block(&cooked, rawBytes)
		}
	} else {
		if sublist > 0 {
			p.inline(&cooked, rawBytes[:sublist])
			p.block(&cooked, rawBytes[sublist:])
		} else {
			p.inline(&cooked, rawBytes)
		}
	}

	cookedBytes := cooked.Bytes()
	parsedEnd := len(cookedBytes)

	for parsedEnd > 0 && cookedBytes[parsedEnd-1] == '\n' {
		parsedEnd--
	}
	p.r.ListItem(out, cookedBytes[:parsedEnd], *flags)

	return line
}

func (p *parser) renderParagraph(out *bytes.Buffer, data []byte) {
	if len(data) == 0 {
		return
	}

	beg := 0
	for data[beg] == ' ' {
		beg++
	}

	end := len(data) - 1

	for end > beg && data[end-1] == ' ' {
		end--
	}

	work := func() bool {
		p.inline(out, data[beg:end])
		return true
	}
	p.r.Paragraph(out, work)
}

func (p *parser) paragraph(out *bytes.Buffer, data []byte) int {
	var prev, line, i int

	for i < len(data) {
		prev = line
		current := data[i:]
		line = i

		if n := p.isEmpty(current); n > 0 {
			if p.flags&EXTENSION_DEFINITION_LISTS != 0 {
				if i < len(data)-1 && data[i+1] == ':' {
					return p.list(out, data[prev:], LIST_TYPE_DEFINITION)
				}
			}

			p.renderParagraph(out, data[:i])
			return i + n
		}

		if i > 0 {
			if level := p.isUnderlinedHeader(current); level > 0 {
				p.renderParagraph(out, data[:prev])

				eol := i - 1
				for prev < eol && data[prev] == ' ' {
					prev++
				}
				for eol > prev && data[eol-1] == ' ' {
					eol--
				}

				work := func(o *bytes.Buffer, pp *parser, d []byte) func() bool {
					return func() bool {
						pp.inline(o, d)
						return true
					}
				}(out, p, data[prev:eol])

				id := ""
				if p.flags&EXTENSION_AUTO_HEADER_IDS != 0 {
					id = SanitizedAnchorName(string(data[prev:eol]))
				}

				p.r.Header(out, work, level, id)

				for data[i] != '\n' {
					i++
				}
				return i
			}
		}

		if p.flags&EXTENSION_LAX_HTML_BLOCKS != 0 {
			if data[i] == '<' && p.html(out, current, false) > 0 {
				p.renderParagraph(out, data[:i])
				return i
			}
		}

		if p.isPrefixHeader(current) || p.isHRule(current) {
			p.renderParagraph(out, data[:i])
			return i
		}

		if p.flags&EXTENSION_FENCED_CODE != 0 {
			if p.fencedCodeBlock(out, current, false) > 0 {
				p.renderParagraph(out, data[:i])
				return i
			}
		}

		if p.flags&EXTENSION_DEFINITION_LISTS != 0 {
			if p.dliPrefix(current) != 0 {
				return p.list(out, data[prev:], LIST_TYPE_DEFINITION)
			}
		}

		if p.flags&EXTENSION_NO_EMPTY_LINE_BEFORE_BLOCK != 0 {
			if p.uliPrefix(current) != 0 ||
				p.oliPrefix(current) != 0 ||
				p.quotePrefix(current) != 0 ||
				p.codePrefix(current) != 0 {
				p.renderParagraph(out, data[:i])
				return i
			}
		}

		for data[i] != '\n' {
			i++
		}
		i++
	}

	p.renderParagraph(out, data[:i])
	return i
}

func SanitizedAnchorName(text string) string {
	var anchorName []rune
	futureDash := false
	for _, r := range text {
		switch {
		case unicode.IsLetter(r) || unicode.IsNumber(r):
			if futureDash && len(anchorName) > 0 {
				anchorName = append(anchorName, '-')
			}
			futureDash = false
			anchorName = append(anchorName, unicode.ToLower(r))
		default:
			futureDash = true
		}
	}
	return string(anchorName)
}
