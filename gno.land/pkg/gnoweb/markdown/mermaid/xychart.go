package mermaid

import (
	"math"
	"strconv"
	"strings"
)

const (
	plotHeight = 280.0
	chartWidth = 700.0 // Mermaid's default width, which the plot fills
	minBand    = 36.0  // width of one category
	tickLen    = 4.0
	axisTitleH = 26.0 // room for the title above the y-axis or below the x-axis
	maxTicks   = 12
	barGroupW  = 0.7 // share of a category the bars take
)

const (
	classGrid class = "gno-mermaid__grid"
	classPlot class = "gno-mermaid__plot"
)

type series struct {
	bar    bool
	values []float64
}

type xychart struct {
	xTitle, yTitle string
	categories     []string
	yMin, yMax     float64 // equal until a range is set
	series         []series
}

func parseXYChart(args string, lines []line, m *meta) (diagram, error) {
	if args != "" {
		return nil, errorf(0, "unsupported xychart option %q", args)
	}
	d := &xychart{}
	for _, l := range lines {
		if t, ok := title(l.text); ok {
			m.title = t
			continue
		}
		word, rest, _ := strings.Cut(l.text, " ")
		rest = strings.TrimSpace(rest)
		switch lowerASCII(word) {
		case "x-axis":
			if err := d.parseXAxis(l, rest); err != nil {
				return nil, err
			}
		case "y-axis":
			if err := d.parseYAxis(l, rest); err != nil {
				return nil, err
			}
		case "bar", "line":
			if !strings.HasPrefix(rest, "[") {
				return nil, errorf(l.n, "unsupported %s statement", word)
			}
			items, err := parseList(l, rest)
			if err != nil {
				return nil, err
			}
			if len(items) == 0 {
				return nil, errorf(l.n, "empty %s series", word)
			}
			s := series{bar: equalFold(word, "bar"), values: make([]float64, len(items))}
			for i, it := range items {
				if s.values[i], err = parseNumber(it); err != nil {
					return nil, errorf(l.n, "%q is not a number", it)
				}
			}
			d.series = append(d.series, s)
		default:
			return nil, errorf(l.n, "unsupported statement %q", word)
		}
	}
	if len(d.series) == 0 {
		return nil, errorf(0, "xychart without data")
	}
	if len(d.categories) > 0 {
		// Mermaid draws as many points as there are categories.
		for i := range d.series {
			s := &d.series[i]
			s.values = s.values[:min(len(s.values), len(d.categories))]
		}
	} else {
		n := 0
		for _, s := range d.series {
			n = max(n, len(s.values))
		}
		for i := range n {
			d.categories = append(d.categories, strconv.Itoa(i+1))
		}
	}
	if d.yMin == d.yMax {
		d.autoRange()
	}
	// Tick labels have at most 15 decimals, and a float64 holds about 16
	// significant digits: a smaller step, or one too small for the
	// magnitude of the range, labels ticks the same.
	step := niceStep(d.yMax - d.yMin)
	if !(step >= 1e-15) || math.IsInf(step, 0) || step < max(math.Abs(d.yMin), math.Abs(d.yMax))*1e-13 {
		return nil, errorf(0, "y range too small")
	}
	return d, nil
}

// parseXAxis parses `x-axis ["title"] [a, b, …]`. A numeric axis is not
// supported.
func (d *xychart) parseXAxis(l line, rest string) error {
	axisTitle, list := splitTitle(rest)
	if strings.Contains(list, "-->") {
		return errorf(l.n, "numeric x-axis")
	}
	d.xTitle = axisTitle
	if list == "" {
		return nil
	}
	cats, err := parseList(l, list)
	d.categories = cats
	return err
}

// parseYAxis parses `y-axis ["title"] [min --> max]`.
func (d *xychart) parseYAxis(l line, rest string) error {
	axisTitle, rng := splitTitle(rest)
	d.yTitle = axisTitle
	if rng == "" {
		return nil
	}
	lo, hi, ok := strings.Cut(rng, "-->")
	yMin, err1 := parseNumber(strings.TrimSpace(lo))
	yMax, err2 := parseNumber(strings.TrimSpace(hi))
	if !ok || err1 != nil || err2 != nil || yMax <= yMin {
		return errorf(l.n, "malformed y-axis range")
	}
	d.yMin, d.yMax = yMin, yMax
	return nil
}

// splitTitle splits an axis statement into its title and what follows: a
// list or a range. An unquoted title runs up to the list, or up to the word
// before the arrow of a range.
func splitTitle(s string) (string, string) {
	if inner, rest, ok := cutQuoted(s); ok {
		return decodeEntities(inner), rest
	}
	end := strings.IndexByte(s, '[')
	if end < 0 {
		end = len(s)
		if before, _, ok := strings.Cut(s, "-->"); ok {
			end = strings.LastIndexByte(strings.TrimSpace(before), ' ') + 1
		}
	}
	return decodeEntities(strings.TrimSpace(s[:end])), strings.TrimSpace(s[end:])
}

// parseList parses `[a, "b, c", d]` into its items.
func parseList(l line, s string) ([]string, error) {
	if !strings.HasPrefix(s, "[") || !strings.HasSuffix(s, "]") {
		return nil, errorf(l.n, "malformed list")
	}
	s = s[1 : len(s)-1]
	var items []string
	for s = strings.TrimSpace(s); s != ""; {
		var item string
		if s[0] == '"' {
			var ok bool
			if item, s, ok = cutQuoted(s); !ok {
				return nil, errorf(l.n, "unclosed quote")
			}
		} else {
			i := strings.Index(s, ",")
			if i < 0 {
				i = len(s)
			}
			item, s = strings.TrimSpace(s[:i]), s[i:]
		}
		items = append(items, decodeEntities(item))
		if s == "" {
			break
		}
		if s[0] != ',' {
			return nil, errorf(l.n, "malformed list")
		}
		s = strings.TrimSpace(s[1:])
	}
	return items, nil
}

// autoRange sets the y range from the data, from zero or below.
func (d *xychart) autoRange() {
	lo, hi := 0.0, 0.0
	for _, s := range d.series {
		for _, v := range s.values {
			lo, hi = min(lo, v), max(hi, v)
		}
	}
	if hi == lo {
		hi = lo + 1
	}
	step := niceStep(hi - lo)
	d.yMin, d.yMax = math.Floor(lo/step)*step, math.Ceil(hi/step)*step
}

// niceStep returns a tick step of 1, 2 or 5 times a power of ten that cuts
// span into about five intervals.
func niceStep(span float64) float64 {
	raw := span / 5
	p := math.Pow(10, math.Floor(math.Log10(raw)))
	for _, f := range []float64{1, 2, 5} {
		if raw <= f*p {
			return f * p
		}
	}
	return 10 * p
}

// ticks returns the y values to label: multiples of the step, about five.
func (d *xychart) ticks() (vals []float64, labels []string) {
	step := niceStep(d.yMax - d.yMin)
	decimals := min(max(0, int(-math.Floor(math.Log10(step)))), 15)
	first := math.Ceil(d.yMin/step-1e-9) * step // -0.6/0.2 is -2.9999999999999996
	for i := range maxTicks + 1 {
		v := first + float64(i)*step
		if v > d.yMax+step/1e6 {
			break
		}
		vals = append(vals, v)
		labels = append(labels, strconv.FormatFloat(v, 'f', decimals, 64))
	}
	return vals, labels
}

func (d *xychart) draw(w *svgWriter) (float64, float64) {
	vals, labels := d.ticks()
	labelW := linesWidth(labels, smallSize, false)
	band := minBand
	for _, c := range d.categories {
		band = max(band, textWidth(c, smallSize, false)+12)
	}
	n := float64(len(d.categories))
	plotW := max(band*n, chartWidth-(pad+labelW+tickLen+6)-pad)
	band = plotW / n

	top := pad
	if d.yTitle != "" {
		top += axisTitleH
	}
	// The x-axis title is centered under the plot, which moves right when
	// the title is wider.
	xTitleW := textWidth(d.xTitle, smallSize, true)
	left := max(pad, pad+(xTitleW-plotW)/2-labelW-tickLen-6) // of the tick labels and the y-axis title
	x0 := left + labelW + tickLen + 6
	width := max(x0+max(plotW, (plotW+xTitleW)/2), left+textWidth(d.yTitle, smallSize, true)) + pad
	bottom := top + plotHeight
	height := bottom + tickLen + 20 + pad
	if d.xTitle != "" {
		height += axisTitleH
	}
	y := func(v float64) float64 {
		v = min(max(v, d.yMin), d.yMax)
		return bottom - (v-d.yMin)/(d.yMax-d.yMin)*plotHeight
	}

	w.text(classTitle, left, top-14, anchorStart, smallSize, d.yTitle)
	for i, v := range vals {
		w.line(classGrid, x0, y(v), x0+plotW, y(v))
		w.text(classMuted, x0-tickLen-4, y(v)+4, anchorEnd, smallSize, labels[i])
	}
	w.line(classStroke, x0, top, x0, bottom)
	w.line(classStroke, x0, bottom, x0+plotW, bottom)
	for i, c := range d.categories {
		cx := x0 + band*(float64(i)+0.5)
		w.line(classStroke, cx, bottom, cx, bottom+tickLen)
		w.text(classMuted, cx, bottom+tickLen+14, anchorMiddle, smallSize, c)
	}
	w.text(classTitle, x0+plotW/2, height-pad-4, anchorMiddle, smallSize, d.xTitle)

	var bars int
	for _, s := range d.series {
		if s.bar {
			bars++
		}
	}
	groupW := band * barGroupW
	barW := groupW / float64(max(bars, 1))
	zero := y(0)
	bar := 0
	for k, s := range d.series {
		if s.bar {
			w.group(classFill + " " + color(k))
			for i, v := range s.values {
				bx := x0 + band*float64(i) + (band-groupW)/2 + barW*float64(bar)
				vy := y(v)
				w.rect("", bx, min(vy, zero), barW, math.Abs(zero-vy), 0)
			}
			w.end()
			bar++
			continue
		}
		pts := make([]float64, 0, 2*len(s.values))
		for i, v := range s.values {
			pts = append(pts, x0+band*(float64(i)+0.5), y(v))
		}
		w.polyline(classPlot+" "+color(k), pts...)
	}
	return width, height
}
