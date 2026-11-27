package mermaid

import "strings"

const (
	columnMaxW = 200.0 // text wraps past it
	columnMinW = 100.0
	columnGap  = 16.0
	cardGap    = 10.0
)

const (
	classPeriod  class = "gno-mermaid__period"
	classCard    class = "gno-mermaid__card"
	classOnColor class = "gno-mermaid__on-color"
)

type period struct {
	label   []string
	events  [][]string
	section int // index into sections, or -1
}

type timeline struct {
	sections [][]string
	periods  []*period
}

func parseTimeline(args string, lines []line, m *meta) (diagram, error) {
	if args != "" {
		return nil, errorf(0, "unsupported timeline option %q", args)
	}
	d := &timeline{}
	for _, l := range lines {
		if t, ok := title(l.text); ok {
			m.title = t
			continue
		}
		if strings.HasPrefix(l.text, "#") {
			continue // a comment, in Mermaid's timeline syntax
		}
		if rest, ok := cutPrefixFold(l.text, "section "); ok {
			// Mermaid ends a section at a colon and reads the rest as an event.
			if strings.Contains(rest, ":") {
				return nil, errorf(l.n, "colon in a section")
			}
			d.sections = append(d.sections, label(rest))
			continue
		}
		// As in Mermaid, a colon followed by a space separates the events, so
		// `10:30` or a URL stays whole.
		parts := strings.Split(l.text, ": ")
		var p *period
		if strings.TrimSpace(parts[0]) == "" {
			if len(d.periods) == 0 {
				return nil, errorf(l.n, "event before any period")
			}
			p = d.periods[len(d.periods)-1]
		} else {
			if strings.Contains(parts[0], ":") {
				return nil, errorf(l.n, "colon in a period")
			}
			p = &period{label: label(parts[0]), section: len(d.sections) - 1}
			d.periods = append(d.periods, p)
		}
		for _, e := range parts[1:] {
			if strings.TrimSpace(e) == "" {
				return nil, errorf(l.n, "empty event")
			}
			p.events = append(p.events, label(e))
		}
	}
	if len(d.periods) == 0 {
		return nil, errorf(0, "timeline without periods")
	}
	return d, nil
}

func (d *timeline) draw(w *svgWriter) (float64, float64) {
	type column struct {
		x, w  float64
		label []string
		cards [][]string
	}
	inner := columnMaxW - 2*cardPad
	cols := make([]column, len(d.periods))
	x := pad
	var periodH float64
	for i, p := range d.periods {
		c := column{label: wrap(p.label, fontSize, inner, true)}
		cw := linesWidth(c.label, fontSize, true)
		for _, e := range p.events {
			lines := wrap(e, fontSize, inner, false)
			c.cards = append(c.cards, lines)
			cw = max(cw, linesWidth(lines, fontSize, false))
		}
		c.x, c.w = x, max(cw+2*cardPad, columnMinW)
		x += c.w + columnGap
		cols[i] = c
		periodH = max(periodH, cardHeight(c.label, fontSize))
	}
	contentW := x - columnGap - pad

	// Sections span the columns of their periods, their labels wrapped to
	// that span.
	type span struct {
		l, r  float64
		label []string
	}
	spans := make([]*span, len(d.sections))
	for i, p := range d.periods {
		if p.section < 0 {
			continue
		}
		s := spans[p.section]
		if s == nil {
			s = &span{l: cols[i].x}
			spans[p.section] = s
		}
		s.r = cols[i].x + cols[i].w
	}
	top := pad
	sectionH := 0.0
	for i, s := range spans {
		if s != nil {
			s.label = wrap(d.sections[i], fontSize, s.r-s.l-2*cardPad, true)
			sectionH = max(sectionH, cardHeight(s.label, fontSize)+cardGap)
		}
	}
	periodTop := top + sectionH
	axisY := periodTop + periodH + 18
	cardsTop := axisY + 18

	bottom := cardsTop
	for _, c := range cols {
		y := cardsTop
		for _, card := range c.cards {
			y += cardHeight(card, fontSize) + cardGap
		}
		bottom = max(bottom, y)
	}
	colorOf := func(i int) class {
		if p := d.periods[i]; p.section >= 0 {
			return color(p.section)
		}
		if len(d.sections) > 0 {
			return color(len(d.sections)) // periods before the first section
		}
		return color(i)
	}

	for i, s := range spans {
		if s == nil {
			continue // a section without periods
		}
		l, r := s.l, s.r
		w.rect(classPeriod+" "+color(i), l, top, r-l, sectionH-cardGap, boxRadius)
		w.text(classOnColor+" "+classTitle, (l+r)/2, baseline(top+cardPad, fontSize), anchorMiddle, 0, s.label...)
	}

	end := pad + contentW
	w.line(classEdge, pad, axisY, end+8, axisY)
	w.polygon(classHead, end+12, axisY, end+2, axisY-5, end+2, axisY+5)
	for i, c := range cols {
		cx := c.x
		mid := cx + c.w/2
		if len(c.cards) > 0 {
			w.line(classDashed, mid, periodTop+periodH, mid, bottom-cardGap)
		}
		col := colorOf(i)
		w.rect(classPeriod+" "+col, cx, periodTop, c.w, periodH, boxRadius)
		w.text(classOnColor+" "+classTitle, mid, baseline(periodTop+(periodH-textBlock(len(c.label), fontSize))/2, fontSize), anchorMiddle, 0, c.label...)
		y := cardsTop
		for _, card := range c.cards {
			h := cardHeight(card, fontSize)
			w.rect(classCard+" "+col, cx, y, c.w, h, boxRadius)
			w.text("", mid, baseline(y+cardPad, fontSize), anchorMiddle, 0, card...)
			y += h + cardGap
		}
	}
	return end + 12 + pad, bottom + pad
}
