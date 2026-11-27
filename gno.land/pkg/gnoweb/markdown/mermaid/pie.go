package mermaid

import (
	"math"
	"strconv"
	"strings"
)

const (
	pieRadius  = 110.0
	legendGap  = 32.0 // pie to legend
	legendRow  = 24.0
	swatchSize = 12.0
)

const classSlice class = "gno-mermaid__slice"

type slice struct {
	label string
	value float64
}

type pie struct {
	showData bool
	slices   []slice
	total    float64
}

func parsePie(args string, lines []line, m *meta) (diagram, error) {
	d := &pie{}
	if rest, ok := cutPrefixFold(args, "showData"); ok {
		d.showData, args = true, strings.TrimSpace(rest)
	}
	if t, ok := title(args); ok {
		m.title = t
	} else if args != "" {
		return nil, errorf(0, "unsupported pie option %q", args)
	}
	seen := map[string]bool{}
	for _, l := range lines {
		if t, ok := title(l.text); ok {
			m.title = t
			continue
		}
		if equalFold(l.text, "showData") {
			d.showData = true
			continue
		}
		s, err := parseSlice(l)
		if err != nil {
			return nil, err
		}
		if seen[s.label] {
			continue // Mermaid keeps the first slice of a label
		}
		seen[s.label] = true
		d.slices = append(d.slices, s)
		d.total += s.value
	}
	if d.total <= 0 || math.IsInf(d.total, 0) {
		return nil, errorf(0, "pie values must add up to a positive number")
	}
	return d, nil
}

// parseSlice parses `"label" : value`.
func parseSlice(l line) (slice, error) {
	if !strings.HasPrefix(l.text, `"`) {
		return slice{}, errorf(l.n, "unsupported statement %q", firstWord(l.text))
	}
	name, rest, ok := cutQuoted(l.text)
	if !ok {
		return slice{}, errorf(l.n, "unclosed quote")
	}
	value, ok := strings.CutPrefix(rest, ":")
	if !ok {
		return slice{}, errorf(l.n, "slice without a value")
	}
	v, err := parseNumber(strings.TrimSpace(value))
	if err != nil || v < 0 {
		return slice{}, errorf(l.n, "slice value must be a non-negative number")
	}
	return slice{label: decodeEntities(name), value: math.Abs(v)}, nil // -0 is 0
}

func percent(part, total float64) string {
	return strconv.FormatFloat(math.Round(part/total*1000)/10, 'f', -1, 64) + "%"
}

func (d *pie) legend(s slice) string {
	text := s.label + " " + percent(s.value, d.total)
	if d.showData {
		format := byte('f')
		if s.value != 0 && s.value < 1e-6 {
			format = 'g' // as JavaScript writes it, not 300 digits
		}
		// JavaScript writes no leading zero in an exponent: 1e-7, not 1e-07.
		text += " (" + strings.Replace(strconv.FormatFloat(s.value, format, -1, 64), "e-0", "e-", 1) + ")"
	}
	return text
}

// sliceColor returns the color of the i-th slice. The last one skips the
// first slice's color when the palette wraps onto it, as the two touch.
func (d *pie) sliceColor(i int) class {
	if n := len(d.slices); i == n-1 && n > 1 && i%len(palette) == 0 {
		return color(i + 1)
	}
	return color(i)
}

func (d *pie) draw(w *svgWriter) (float64, float64) {
	var legendW float64
	for _, s := range d.slices {
		legendW = max(legendW, swatchSize+8+textWidth(d.legend(s), fontSize, false))
	}
	legendH := float64(len(d.slices)) * legendRow
	bodyH := max(2*pieRadius, legendH)
	cx, cy := pad+pieRadius, pad+bodyH/2
	angle := -math.Pi / 2 // from the top, clockwise
	for i, s := range d.slices {
		if s.value == 0 {
			continue
		}
		c := classSlice + " " + d.sliceColor(i)
		if s.value == d.total {
			w.circle(c, cx, cy, pieRadius)
			break
		}
		sweep := s.value / d.total * 2 * math.Pi
		at := func(a float64) (float64, float64) {
			return cx + pieRadius*math.Cos(a), cy + pieRadius*math.Sin(a)
		}
		x0, y0 := at(angle)
		p := new(pathData).cmd('M', cx, cy).cmd('L', x0, y0)
		if sweep > math.Pi {
			// A slice close to a full turn would end where it starts once
			// rounded, and SVG drops such an arc: go through the middle.
			xm, ym := at(angle + sweep/2)
			p.cmd('A', pieRadius, pieRadius, 0, 0, 1, xm, ym)
		}
		angle += sweep
		x1, y1 := at(angle)
		w.path(c, p.cmd('A', pieRadius, pieRadius, 0, 0, 1, x1, y1).cmd('Z'))
	}

	lx := pad + 2*pieRadius + legendGap
	ly := cy - legendH/2
	for i, s := range d.slices {
		y := ly + float64(i)*legendRow
		w.rect(classFill+" "+d.sliceColor(i), lx, y+(legendRow-swatchSize)/2, swatchSize, swatchSize, 2)
		w.text("", lx+swatchSize+8, y+legendRow/2+5, anchorStart, 0, d.legend(s))
	}
	return lx + legendW + pad, pad + bodyH + pad
}
