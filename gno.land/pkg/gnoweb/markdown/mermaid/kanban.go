package mermaid

import (
	"strconv"
	"strings"
)

const (
	laneWidth  = 220.0
	laneGap    = 12.0
	laneHeader = 36.0
	stripeW    = 4.0
)

const classTicket class = "gno-mermaid__ticket"

// priorities maps a card priority to the series color of its stripe.
var priorities = map[string]int{"very high": 4, "high": 2, "low": 1, "very low": 6}

type kanbanCard struct {
	label    []string
	details  string // ticket and assignee
	priority int    // a series color, or -1
}

type kanbanColumn struct {
	label []string
	cards []kanbanCard
}

type kanban struct {
	columns []*kanbanColumn
}

func parseKanban(args string, lines []line, _ *meta) (diagram, error) {
	if args != "" {
		return nil, errorf(0, "unsupported kanban option %q", args)
	}
	d := &kanban{}
	colIndent := -1
	for _, l := range lines {
		card, err := parseKanbanNode(l)
		if err != nil {
			return nil, err
		}
		if colIndent < 0 {
			colIndent = l.indent // as in Mermaid, the first line sets the column level
		}
		if l.indent < colIndent {
			return nil, errorf(l.n, "line indented less than the columns")
		}
		if l.indent == colIndent {
			if card.details != "" || card.priority >= 0 {
				return nil, errorf(l.n, "metadata on a column")
			}
			d.columns = append(d.columns, &kanbanColumn{label: card.label})
			continue
		}
		c := d.columns[len(d.columns)-1]
		c.cards = append(c.cards, card)
	}
	if len(d.columns) == 0 {
		return nil, errorf(0, "kanban without columns")
	}
	return d, nil
}

// parseKanbanNode parses `id[label]@{ key: value, … }`, where only the id is
// required. The metadata drawn are label, ticket, assigned and priority;
// other keys are ignored. The other node shapes, icons and classes are not
// supported.
func parseKanbanNode(l line) (kanbanCard, error) {
	card := kanbanCard{priority: -1}
	node, metadata, hasMeta := strings.Cut(l.text, "@{")
	node = strings.TrimSpace(node)
	if strings.HasPrefix(node, "::") || strings.Contains(node, ":::") {
		return card, errorf(l.n, "kanban icons and classes are not supported")
	}
	text := node
	if i := strings.Index(node, "["); i >= 0 {
		if !strings.HasSuffix(node, "]") || strings.HasPrefix(node[i:], "[[") {
			return card, errorf(l.n, "malformed label")
		}
		text = unquote(node[i+1 : len(node)-1])
	} else if strings.ContainsAny(node, "(){}") {
		return card, errorf(l.n, "kanban node shapes are not supported")
	}
	if hasMeta {
		body, ok := strings.CutSuffix(strings.TrimSpace(metadata), "}")
		if !ok {
			return card, errorf(l.n, "unclosed metadata")
		}
		pairs, ok := splitMetadata(body)
		if !ok {
			return card, errorf(l.n, "unclosed quote")
		}
		var details []string
		for _, kv := range pairs {
			k, v, ok := strings.Cut(kv, ":")
			if !ok {
				return card, errorf(l.n, "malformed metadata")
			}
			switch strings.TrimSpace(k) {
			case "label":
				text = unquote(strings.TrimSpace(v)) // label decodes it
			case "ticket", "assigned":
				if v = plain(v); v != "" {
					details = append(details, v)
				}
			case "priority":
				if p, ok := priorities[lowerASCII(plain(v))]; ok {
					card.priority = p
				}
			}
		}
		card.details = strings.Join(details, " · ")
	}
	if card.label = label(text); card.label == nil {
		return card, errorf(l.n, "empty label")
	}
	return card, nil
}

// splitMetadata splits `key: value, …` at the commas outside quotes, as
// Mermaid's YAML reading of it does: as in YAML, a quote opens a quoted key
// or value only as its first character, so O'Brien is plain. ok is false for
// an unclosed quote.
func splitMetadata(s string) (pairs []string, ok bool) {
	pairs = make([]string, 0, 4) // a card rarely has more pairs
	var quote byte
	start := 0
	value := false // past the colon of the pair
	blank := true  // nothing but spaces yet in the key or the value
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0: // jump to the closing quote
			j := strings.IndexByte(s[i:], quote)
			if j < 0 {
				return append(pairs, s[start:]), false
			}
			i, quote = i+j, 0
		case !metadataByte[c]:
			blank = false
		case c == ',':
			pairs, start = append(pairs, s[start:i]), i+1
			value, blank = false, true
		case c == ':' && !value:
			value, blank = true, true
		case c == ' ': // leaves blank as it is
		case blank && c != ':':
			quote, blank = c, false
		default:
			blank = false
		}
	}
	return append(pairs, s[start:]), quote == 0
}

// metadataByte marks the bytes splitMetadata acts on.
var metadataByte = [256]bool{',': true, ':': true, ' ': true, '"': true, '\'': true}

func (d *kanban) draw(w *svgWriter) (float64, float64) {
	inner := laneWidth - 2*laneGap - 2*cardPad
	type box struct {
		lines   []string
		details []string
		h       float64
	}
	lanes := make([][]box, len(d.columns))
	titles := make([][]string, len(d.columns))
	header := laneHeader
	for i, col := range d.columns {
		titles[i] = wrap(col.label, fontSize, inner-24, true)
		header = max(header, textBlock(len(titles[i]), fontSize)+20)
	}
	var laneH float64
	for i, col := range d.columns {
		h := header
		for _, c := range col.cards {
			k := box{lines: wrap(c.label, fontSize, inner, false)}
			if c.details != "" {
				k.details = wrap([]string{c.details}, smallSize, inner, false)
			}
			k.h = cardHeight(k.lines)
			if len(k.details) > 0 {
				k.h += textBlock(len(k.details), smallSize) + 4
			}
			lanes[i] = append(lanes[i], k)
			h += k.h + laneGap
		}
		laneH = max(laneH, h+laneGap/2)
	}
	for i, col := range d.columns {
		x := pad + float64(i)*(laneWidth+laneGap)
		w.rect(classBox, x, pad, laneWidth, laneH, 6)
		w.text(classTitle, x+laneGap, baseline(pad+10, fontSize), anchorStart, 0, titles[i]...)
		w.text(classMuted, x+laneWidth-laneGap, baseline(pad+11, smallSize), anchorEnd, smallSize, strconv.Itoa(len(col.cards)))
		y := pad + header
		for j, k := range lanes[i] {
			cx, cw := x+laneGap, laneWidth-2*laneGap
			w.rect(classTicket, cx, y, cw, k.h, boxRadius)
			if p := col.cards[j].priority; p >= 0 {
				w.rect(classFill+" "+color(p), cx, y, stripeW, k.h, 2)
			}
			w.text("", cx+cardPad, baseline(y+cardPad, fontSize), anchorStart, 0, k.lines...)
			if len(k.details) > 0 {
				ty := y + cardPad + textBlock(len(k.lines), fontSize) + 4
				w.text(classMuted, cx+cardPad, baseline(ty, smallSize), anchorStart, smallSize, k.details...)
			}
			y += k.h + laneGap
		}
	}
	return pad + float64(len(d.columns))*(laneWidth+laneGap) - laneGap + pad, pad + laneH + pad
}
