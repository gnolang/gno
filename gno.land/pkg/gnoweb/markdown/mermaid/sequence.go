package mermaid

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// Sequence diagram limits. Spacing the participants takes time in the
// number of messages times the number of participants, which the first
// bounds.
const (
	maxParticipants = 50
	maxAutonumber   = 1_000_000 // start and step
)

// Sequence diagram geometry, in px.
const (
	actorPadX     = 14.0 // text to box side
	actorMinWidth = 72.0
	actorMinH     = 40.0
	actorGap      = 80.0 // between neighboring actor boxes
	figureHeight  = 38.0 // a stick figure, above its label
	messageGap    = 18.0 // label end to lifeline
	rowGap        = 24.0 // below a message or a note, to the next row
	noteGap       = 12.0 // note to lifeline
	notePad       = 10.0
	selfLoopW     = 34.0
	selfLoopH     = 18.0
	frameInset    = 10.0 // frame side to what it holds
	tabHeight     = 20.0
	badgeH        = 18.0 // an autonumber
	badgeSize     = 11.0
)

const (
	classEdgeDash  class = "gno-mermaid__edge gno-mermaid__dashed"
	classNote      class = "gno-mermaid__note"
	classFrame     class = "gno-mermaid__frame"
	classFrameLine class = "gno-mermaid__frame gno-mermaid__dashed"
	classOnBadge   class = "gno-mermaid__on-badge"
)

type seqParticipant struct {
	label []string
	actor bool // a stick figure rather than a box
	w, h  float64
	x     float64
}

type seqArrow struct {
	dashed bool
	head   byte // 0 for none, '>' filled, 'x' a cross, ')' open
	both   bool // a head at each end
}

// seqArrows are the message arrows, a longer one before any it starts with.
var seqArrows = []struct {
	tok string
	a   seqArrow
}{
	{"<<-->>", seqArrow{dashed: true, head: '>', both: true}},
	{"<<->>", seqArrow{head: '>', both: true}},
	{"-->>", seqArrow{dashed: true, head: '>'}},
	{"->>", seqArrow{head: '>'}},
	{"--x", seqArrow{dashed: true, head: 'x'}},
	{"-x", seqArrow{head: 'x'}},
	{"--)", seqArrow{dashed: true, head: ')'}},
	{"-)", seqArrow{head: ')'}},
	{"-->", seqArrow{dashed: true}},
	{"->", seqArrow{}},
}

type seqKind int

const (
	seqMessage seqKind = iota
	seqNote
	seqOpen    // loop, alt, opt, par, critical, break
	seqSection // else, and, option
	seqClose   // end
)

const (
	noteLeft = iota
	noteRight
	noteOver
)

type seqEvent struct {
	kind     seqKind
	from, to int // participants; a note over one has from == to
	arrow    seqArrow
	text     []string
	num      string // the autonumber, or ""
	notePos  int
	block    *seqBlock // of an open event
	keyword  string    // of an open event
	n        int       // source line
	y, h     float64
	l, r     float64 // horizontal extent, relative to participant 0
}

type seqBlock struct {
	open     *seqEvent
	sections []*seqEvent
	parent   *seqBlock
	l, r     float64 // extent of what it holds; l > r while it holds nothing
	top, end float64
}

type sequence struct {
	parts  []*seqParticipant
	events []*seqEvent
	blocks []*seqBlock
}

// sectionOf maps a section keyword to the block it divides.
var sectionOf = map[string]string{"else": "alt", "and": "par", "option": "critical"}

func parseSequence(args string, lines []line, m *meta) (diagram, error) {
	if args != "" {
		return nil, errorf(0, "statement on the sequenceDiagram line")
	}
	d := &sequence{}
	ids := map[string]int{}
	participant := func(n int, id string) (int, error) {
		if !validName(id) {
			return 0, errorf(n, "unsupported participant name %q", id)
		}
		if i, ok := ids[id]; ok {
			return i, nil
		}
		if len(d.parts) == maxParticipants {
			return 0, errorf(n, "more than %d participants", maxParticipants)
		}
		ids[id] = len(d.parts)
		d.parts = append(d.parts, &seqParticipant{label: []string{id}})
		return len(d.parts) - 1, nil
	}
	var open *seqBlock
	// Mermaid counts every message, numbered or not, and a bare autonumber
	// resumes the count.
	numbering, next, step := false, 1, 1
	stmts, err := statements(lines)
	if err != nil {
		return nil, err
	}
	for _, l := range stmts {
		if t, ok := title(l.text); ok {
			m.title = t
			continue
		}
		word, rest, _ := strings.Cut(l.text, " ")
		rest = strings.TrimSpace(rest)
		switch kw := lowerASCII(word); kw {
		case "participant", "actor":
			id, alias, hasAlias := rest, "", false
			if i := strings.Index(lowerASCII(rest), " as "); i >= 0 {
				id, alias, hasAlias = rest[:i], rest[i+4:], true
			}
			id = strings.TrimSpace(id)
			_, declared := ids[id]
			i, err := participant(l.n, id)
			if err != nil {
				return nil, err
			}
			if declared && !hasAlias {
				break // as in Mermaid, a participant declared again keeps its shape
			}
			p := d.parts[i]
			p.actor = kw == "actor"
			if hasAlias {
				p.label = seqLabel(alias)
			}
		case "note":
			e, err := parseNote(l, rest, participant)
			if err != nil {
				return nil, err
			}
			d.events = append(d.events, e)
		case "loop", "alt", "opt", "par", "critical", "break":
			b := &seqBlock{parent: open}
			e := &seqEvent{kind: seqOpen, keyword: kw, text: seqLabel(rest), block: b, n: l.n}
			b.open = e
			open = b
			d.blocks = append(d.blocks, b)
			d.events = append(d.events, e)
		case "else", "and", "option":
			if open == nil || open.open.keyword != sectionOf[kw] {
				return nil, errorf(l.n, "%q outside its block", kw)
			}
			e := &seqEvent{kind: seqSection, text: seqLabel(rest)}
			open.sections = append(open.sections, e)
			d.events = append(d.events, e)
		case "end":
			if open == nil || rest != "" {
				return nil, errorf(l.n, "unexpected end")
			}
			d.events = append(d.events, &seqEvent{kind: seqClose})
			open = open.parent
		case "autonumber":
			numbering = true
			f := strings.Fields(rest)
			if len(f) == 1 && equalFold(f[0], "off") {
				numbering = false
				break
			}
			if len(f) > 2 {
				return nil, errorf(l.n, "malformed autonumber")
			}
			// As in Mermaid: a start alone sets the step to 1, and a start or
			// a step of 0 keeps the current one.
			if len(f) == 1 {
				step = 1
			}
			for i, v := range f {
				n, err := strconv.Atoi(v)
				if err != nil || n < 0 || n > maxAutonumber {
					return nil, errorf(l.n, "malformed autonumber")
				}
				switch {
				case n == 0:
				case i == 0:
					next = n
				default:
					step = n
				}
			}
		case "activate", "deactivate":
			// Activation bars are not drawn, but Mermaid refuses an unknown
			// participant.
			if _, ok := ids[rest]; !ok {
				return nil, errorf(l.n, "%s of an unknown participant %q", kw, rest)
			}
		default:
			e, err := parseMessage(l, participant)
			if err != nil {
				return nil, err
			}
			if numbering {
				e.num = strconv.Itoa(next)
			}
			next += step
			d.events = append(d.events, e)
		}
	}
	if open != nil {
		return nil, errorf(open.open.n, "%s block is not closed", open.open.keyword)
	}
	if len(d.parts) == 0 {
		return nil, errorf(0, "no participants")
	}
	return d, nil
}

// statements splits lines at the semicolons that separate statements in
// Mermaid, all but those that end an entity code such as #59;, and drops the
// comments (see hashComment): a line, or its last statement, starting with a
// # that starts no entity code. Mermaid also ends a line at such a #
// anywhere else, where it stays text here; a semicolon after it is refused,
// since Mermaid would not read the statement that follows.
func statements(lines []line) ([]line, error) {
	var out []line
	for _, l := range lines {
		if hashComment(l.text) {
			continue
		}
		start, hash := 0, false
		for i := 0; i < len(l.text); i++ {
			switch l.text[i] {
			case '#':
				if end := entityEnd(l.text, i); end > 0 {
					i = end - 1
				} else {
					hash = true
				}
			case ';':
				if hash {
					return nil, errorf(l.n, `"#" before ";"`)
				}
				if s := strings.TrimSpace(l.text[start:i]); s != "" {
					out = append(out, line{n: l.n, text: s})
				}
				start = i + 1
			}
		}
		if s := strings.TrimSpace(l.text[start:]); s != "" && !hashComment(s) {
			out = append(out, line{n: l.n, text: s})
		}
	}
	return out, nil
}

// wrapPrefix matches the wrap: or nowrap: prefix of a label, which Mermaid
// reads as a setting rather than text.
var wrapPrefix = regexp.MustCompile(`^:?(?:no)?wrap:`)

// seqLabel returns the lines of a label as label does, without its wrap: or
// nowrap: prefix.
func seqLabel(s string) []string {
	s = strings.TrimSpace(s)
	return label(strings.TrimPrefix(s, wrapPrefix.FindString(s)))
}

// validName reports whether s can name a participant: Mermaid ends a name
// at a comma, a colon or an arrow, so a name holding one would be read
// differently by Mermaid.
func validName(s string) bool {
	if s == "" || strings.ContainsAny(s, ",:;<>{}") {
		return false
	}
	for _, a := range seqArrows {
		if strings.Contains(s, a.tok) {
			return false
		}
	}
	return true
}

func parseMessage(l line, participant func(int, string) (int, error)) (*seqEvent, error) {
	head, text, ok := strings.Cut(l.text, ":")
	if !ok {
		return nil, errorf(l.n, "unsupported statement %q", firstWord(l.text))
	}
	for i := 0; i < len(head); i++ {
		for _, a := range seqArrows {
			if !strings.HasPrefix(head[i:], a.tok) {
				continue
			}
			from := strings.TrimSpace(head[:i])
			to := strings.TrimSpace(strings.TrimLeft(head[i+len(a.tok):], "+- "))
			// Mermaid's central connection, Alice->>()John, is not drawn, and
			// Mermaid reads a () anywhere in a name as one.
			if strings.Contains(from, "()") || strings.Contains(to, "()") {
				return nil, errorf(l.n, "unsupported central connection")
			}
			f, err := participant(l.n, from)
			if err != nil {
				return nil, err
			}
			t, err := participant(l.n, to)
			if err != nil {
				return nil, err
			}
			return &seqEvent{kind: seqMessage, from: f, to: t, arrow: a.a, text: seqLabel(text)}, nil
		}
	}
	return nil, errorf(l.n, "unsupported statement %q", firstWord(l.text))
}

func parseNote(l line, rest string, participant func(int, string) (int, error)) (*seqEvent, error) {
	where, text, ok := strings.Cut(rest, ":")
	if !ok {
		return nil, errorf(l.n, "note without a colon")
	}
	e := &seqEvent{kind: seqNote, text: seqLabel(text)}
	pos, who, _ := strings.Cut(strings.TrimSpace(where), " ")
	switch lowerASCII(pos) {
	case "left", "right":
		e.notePos = noteLeft
		if equalFold(pos, "right") {
			e.notePos = noteRight
		}
		of, name, _ := strings.Cut(strings.TrimSpace(who), " ")
		if !equalFold(of, "of") {
			return nil, errorf(l.n, "note %s without of", pos)
		}
		who = name
	case "over":
		e.notePos = noteOver
	default:
		return nil, errorf(l.n, "note position %q", pos)
	}
	names := strings.Split(who, ",")
	if len(names) > 2 || len(names) == 2 && e.notePos != noteOver {
		return nil, errorf(l.n, "note over too many participants")
	}
	var err error
	if e.from, err = participant(l.n, strings.TrimSpace(names[0])); err != nil {
		return nil, err
	}
	e.to = e.from
	if len(names) == 2 {
		if e.to, err = participant(l.n, strings.TrimSpace(names[1])); err != nil {
			return nil, err
		}
	}
	return e, nil
}

func badgeWidth(num string) float64 {
	return max(badgeH, textWidth(num, badgeSize, true)+10)
}

// badgeX returns the center of a message's autonumber badge: on the
// sender's lifeline, or past its arrowhead when the arrow has one there,
// on the top of a self message's loop.
func badgeX(e *seqEvent, a, b float64) float64 {
	if !e.arrow.both {
		return a
	}
	dir := 1.0
	if b < a {
		dir = -1
	}
	return a + dir*(12+badgeWidth(e.num)/2)
}

// selfLabelX returns how far right of its lifeline the label of a self
// message starts: 4 px past its autonumber badge when the badge would reach
// under the label, as a two-headed loop's or a long number's does.
func selfLabelX(e *seqEvent) float64 {
	if e.num == "" {
		return messageGap
	}
	return max(messageGap, badgeX(e, 0, 0)+badgeWidth(e.num)/2+4)
}

// layout places the participants and the events and returns the size of
// the drawing and the top of the participants drawn at the bottom.
func (d *sequence) layout() (width, height, footTop float64) {
	parts := d.parts
	var headH float64
	for _, p := range parts {
		tw := linesWidth(p.label, fontSize, false)
		if p.actor {
			p.w = max(tw+8, 40)
			p.h = figureHeight + 4 + textBlock(len(p.label), fontSize)
		} else {
			p.w = max(tw+2*actorPadX, actorMinWidth)
			p.h = max(actorMinH, textBlock(len(p.label), fontSize)+20)
		}
		headH = max(headH, p.h)
	}

	// Center distances, widened until every label fits.
	gaps := make([]float64, len(parts))
	for i := 1; i < len(parts); i++ {
		gaps[i] = parts[i-1].w/2 + parts[i].w/2 + actorGap
	}
	need := func(a, b int, dist float64) {
		if a > b {
			a, b = b, a
		}
		if a == b || a < 0 || b >= len(parts) {
			return
		}
		cur := 0.0
		for i := a + 1; i <= b; i++ {
			cur += gaps[i]
		}
		if cur < dist {
			gaps[b] += dist - cur
		}
	}
	for _, e := range d.events {
		tw := linesWidth(e.text, fontSize, false)
		switch e.kind {
		case seqMessage:
			badge := 0.0
			if e.num != "" {
				badge = badgeWidth(e.num)
				if e.arrow.both {
					badge += 24
				}
			}
			if e.from != e.to {
				need(e.from, e.to, tw+2*messageGap+badge)
			} else if e.from+1 < len(parts) {
				need(e.from, e.from+1, max(selfLoopW, selfLabelX(e)+tw)+messageGap+parts[e.from+1].w/2)
			}
		case seqNote:
			w := tw + 2*notePad
			switch {
			case e.notePos == noteRight:
				need(e.from, e.from+1, w+2*noteGap)
			case e.notePos == noteLeft:
				need(e.from-1, e.from, w+2*noteGap)
			case e.from != e.to:
				need(e.from, e.to, w-4*noteGap)
			default: // over one participant: clear of its neighbors' lifelines
				need(e.from-1, e.from, w/2+noteGap)
				need(e.from, e.from+1, w/2+noteGap)
			}
		}
	}
	x := 0.0
	for i, p := range parts {
		x += gaps[i]
		p.x = x
	}

	// Rows, top to bottom, and the horizontal extent of each event.
	y := pad + headH + rowGap
	top := y
	condLine := smallSize * leading
	var cur *seqBlock // the innermost open block
	extend := func(b *seqBlock, l, r float64) { b.l, b.r = min(b.l, l), max(b.r, r) }
	for _, e := range d.events {
		e.y = y
		tb := textBlock(len(e.text), fontSize)
		tw := linesWidth(e.text, fontSize, false)
		switch e.kind {
		case seqMessage:
			a, b := parts[e.from].x, parts[e.to].x
			if e.from == e.to {
				e.h = tb + 6 + selfLoopH + rowGap
				e.l, e.r = a, a+max(selfLoopW+4, selfLabelX(e)+tw)
			} else {
				e.h = tb + 6 + rowGap
				e.l, e.r = min(a, b), max(a, b)
			}
			if e.num != "" {
				bx, bw := badgeX(e, a, b), badgeWidth(e.num)
				e.l, e.r = min(e.l, bx-bw/2), max(e.r, bx+bw/2)
			}
		case seqNote:
			w, x := tw+2*notePad, parts[e.from].x
			e.h = tb + 2*notePad + rowGap
			switch e.notePos {
			case noteLeft:
				e.l, e.r = x-noteGap-w, x-noteGap
			case noteRight:
				e.l, e.r = x+noteGap, x+noteGap+w
			default:
				lo, hi := min(x, parts[e.to].x), max(x, parts[e.to].x)
				if e.from != e.to {
					lo, hi = lo-2*noteGap, hi+2*noteGap
				}
				if hi-lo < w {
					c := (lo + hi) / 2
					lo, hi = c-w/2, c+w/2
				}
				e.l, e.r = lo, hi
			}
		case seqOpen:
			e.h = max(tabHeight, 14+float64(max(len(e.text), 1)-1)*condLine+6) + 8
			cur = e.block
			cur.top, cur.l, cur.r = y, math.Inf(1), math.Inf(-1)
		case seqSection:
			e.h = max(tabHeight, 15+float64(max(len(e.text), 1)-1)*condLine+5) + 4
		case seqClose:
			e.h = 12
			b := cur
			cur = b.parent
			b.end = y + 2
			if b.l > b.r { // it holds nothing
				b.l, b.r = parts[0].x-20, parts[0].x+20
			}
			b.l -= frameInset
			b.r += frameInset
			// Room for the keyword tab and the labels.
			if need := b.labelWidth(); b.r-b.l < need {
				b.r = b.l + need
			}
			if cur != nil {
				extend(cur, b.l, b.r)
			}
		}
		if (e.kind == seqMessage || e.kind == seqNote) && cur != nil {
			extend(cur, e.l, e.r)
		}
		y += e.h
	}
	footTop = max(y, top+8) + 4

	// Extent of the whole drawing, shifted right of the margin.
	l, r := math.Inf(1), math.Inf(-1)
	for _, p := range parts {
		l, r = min(l, p.x-p.w/2), max(r, p.x+p.w/2)
	}
	for _, e := range d.events {
		if e.kind == seqMessage || e.kind == seqNote {
			l, r = min(l, e.l), max(r, e.r)
		}
	}
	for _, b := range d.blocks {
		l, r = min(l, b.l), max(r, b.r)
	}
	shift := pad - l
	for _, p := range parts {
		p.x += shift
	}
	for _, e := range d.events {
		e.l += shift
		e.r += shift
	}
	for _, b := range d.blocks {
		b.l += shift
		b.r += shift
	}
	return r - l + 2*pad, footTop + headH + pad, footTop
}

func (b *seqBlock) labelWidth() float64 {
	w := tabWidth(b.open.keyword) + 8 + linesWidth(bracket(b.open.text), smallSize, false) + frameInset
	for _, s := range b.sections {
		w = max(w, 8+linesWidth(bracket(s.text), smallSize, false)+frameInset)
	}
	return w
}

func tabWidth(keyword string) float64 { return textWidth(keyword, smallSize, true) + 16 }

// bracket puts a block condition in brackets, as Mermaid shows it.
func bracket(lines []string) []string {
	if len(lines) == 0 {
		return nil
	}
	out := append([]string(nil), lines...)
	out[0] = "[" + out[0]
	out[len(out)-1] += "]"
	return out
}

func (d *sequence) draw(w *svgWriter) (float64, float64) {
	width, height, footTop := d.layout()
	for _, p := range d.parts {
		w.line(classDashed, p.x, pad+p.h, p.x, footTop)
	}
	for _, b := range d.blocks {
		drawFrame(w, b)
	}
	for _, e := range d.events {
		switch e.kind {
		case seqNote:
			tb := textBlock(len(e.text), fontSize)
			w.rect(classNote, e.l, e.y, e.r-e.l, tb+2*notePad, boxRadius)
			w.text("", (e.l+e.r)/2, baseline(e.y+notePad, fontSize), anchorMiddle, 0, e.text...)
		case seqMessage:
			d.drawMessage(w, e)
		}
	}
	for _, p := range d.parts {
		drawParticipant(w, p, pad)
		drawParticipant(w, p, footTop)
	}
	return width, height
}

func drawFrame(w *svgWriter, b *seqBlock) {
	w.rect(classFrame, b.l, b.top, b.r-b.l, b.end-b.top, 0)
	tw := tabWidth(b.open.keyword)
	w.path(classBox, new(pathData).
		cmd('M', b.l, b.top).cmd('H', b.l+tw).cmd('V', b.top+tabHeight-6).
		cmd('L', b.l+tw-6, b.top+tabHeight).cmd('H', b.l).cmd('Z'))
	w.text(classTitle, b.l+8, b.top+14, anchorStart, smallSize, b.open.keyword)
	w.text(classMuted, b.l+tw+8, b.top+14, anchorStart, smallSize, bracket(b.open.text)...)
	for _, s := range b.sections {
		w.line(classFrameLine, b.l, s.y, b.r, s.y)
		w.text(classMuted, b.l+8, s.y+15, anchorStart, smallSize, bracket(s.text)...)
	}
}

func (d *sequence) drawMessage(w *svgWriter, e *seqEvent) {
	tb := textBlock(len(e.text), fontSize)
	ly := e.y + tb + 6
	c := classEdge
	if e.arrow.dashed {
		c = classEdgeDash
	}
	a, b := d.parts[e.from].x, d.parts[e.to].x
	if e.from == e.to {
		w.text("", a+selfLabelX(e), baseline(e.y, fontSize), anchorStart, 0, e.text...)
		w.polyline(c, a, ly, a+selfLoopW, ly, a+selfLoopW, ly+selfLoopH, a, ly+selfLoopH)
		drawHead(w, e.arrow.head, a, ly+selfLoopH, -1)
		if e.arrow.both {
			drawHead(w, e.arrow.head, a, ly, -1)
		}
	} else {
		w.text("", (a+b)/2, baseline(e.y, fontSize), anchorMiddle, 0, e.text...)
		dir := 1.0
		if b < a {
			dir = -1
		}
		w.line(c, a, ly, b, ly)
		drawHead(w, e.arrow.head, b, ly, dir)
		if e.arrow.both {
			drawHead(w, e.arrow.head, a, ly, -dir)
		}
	}
	if e.num != "" {
		bx, bw := badgeX(e, a, b), badgeWidth(e.num)
		w.rect(classHead, bx-bw/2, ly-badgeH/2, bw, badgeH, badgeH/2)
		w.text(classOnBadge, bx, ly+4, anchorMiddle, badgeSize, e.num)
	}
}

// drawHead draws an arrowhead with its tip at x, y, pointing right when dir
// is 1 and left when it is -1.
func drawHead(w *svgWriter, head byte, x, y, dir float64) {
	switch head {
	case '>':
		w.polygon(classHead, x, y, x-dir*10, y-5, x-dir*10, y+5)
	case ')':
		w.polyline(classEdge, x-dir*9, y-5, x, y, x-dir*9, y+5)
	case 'x':
		w.line(classEdge, x-dir*11, y-5, x-dir*1, y+5)
		w.line(classEdge, x-dir*11, y+5, x-dir*1, y-5)
	}
}

func drawParticipant(w *svgWriter, p *seqParticipant, top float64) {
	if !p.actor {
		w.rect(classBox, p.x-p.w/2, top, p.w, p.h, boxRadius)
		tb := textBlock(len(p.label), fontSize)
		w.text("", p.x, baseline(top+(p.h-tb)/2, fontSize), anchorMiddle, 0, p.label...)
		return
	}
	x := p.x
	w.circle(classEdge, x, top+8, 7)
	w.line(classEdge, x, top+15, x, top+28)
	w.line(classEdge, x-10, top+20, x+10, top+20)
	w.polyline(classEdge, x-8, top+38, x, top+28, x+8, top+38)
	w.text("", x, baseline(top+figureHeight+4, fontSize), anchorMiddle, 0, p.label...)
}
