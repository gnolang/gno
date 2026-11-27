package mermaid

import (
	"bytes"
	"encoding/xml"
	"math"
	"strconv"
)

// Text sizes and spacing, in px: leading spaces the lines of a text, in
// ems; pad is the margin around a diagram, cardPad the padding inside a box
// of text and boxRadius its corner radius, the same in every diagram type.
const (
	fontSize  = 14.0
	smallSize = 12.0
	titleSize = 16.0
	leading   = 1.3
	pad       = 16.0
	cardPad   = 8.0
	boxRadius = 4.0
)

// class is a class attribute value. Its values are the Go constants below
// and in each diagram file, and the palette, so no text from a source can
// become one. Each class appears whole in this package's sources, which the
// stylesheet's purge scans.
type class string

const (
	classTitle  class = "gno-mermaid__title"
	classMuted  class = "gno-mermaid__muted"
	classBox    class = "gno-mermaid__box"
	classStroke class = "gno-mermaid__stroke"
	classDashed class = "gno-mermaid__stroke gno-mermaid__dashed"
	classEdge   class = "gno-mermaid__edge"
	classHead   class = "gno-mermaid__head"
	classFill   class = "gno-mermaid__fill" // in the series color
)

// palette are the series color classes, in the order series take them.
var palette = [...]class{
	"gno-mermaid__c0", "gno-mermaid__c1", "gno-mermaid__c2", "gno-mermaid__c3",
	"gno-mermaid__c4", "gno-mermaid__c5", "gno-mermaid__c6", "gno-mermaid__c7",
	"gno-mermaid__c8", "gno-mermaid__c9", "gno-mermaid__c10", "gno-mermaid__c11",
}

func color(i int) class { return palette[i%len(palette)] }

func titleHeight(title string) float64 {
	if title == "" {
		return 0
	}
	return titleSize * 2
}

// titleWidth includes the margins.
func titleWidth(title string) float64 {
	if title == "" {
		return 0
	}
	return textWidth(title, titleSize, true) + 2*pad
}

// textBlock returns the height of n lines of text set at size.
func textBlock(n int, size float64) float64 {
	if n == 0 {
		return 0
	}
	return float64(n-1)*size*leading + size*1.2
}

// baseline returns the first baseline of a text block whose top is at y.
func baseline(y, size float64) float64 { return y + size*0.95 }

func cardHeight(lines []string) float64 {
	return textBlock(max(len(lines), 1), fontSize) + 2*cardPad
}

// anchor is a text-anchor value.
type anchor string

const (
	anchorStart  anchor = "start"
	anchorMiddle anchor = "middle"
	anchorEnd    anchor = "end"
)

// svgWriter writes the elements of a diagram. Its methods are the only way
// anything gets into the SVG, so they are its allowlist: the elements are
// svg, desc, g, rect, line, path, polygon, polyline, circle, text and tspan;
// attribute values are numbers it formats (a translation included), classes
// and anchors, or escaped text. No method writes a style, an event handler,
// a link or a reference to another resource.
//
// Once it holds more than MaxOutputLen bytes, it drops every element it is
// given: Render refuses such a drawing, so one too large stops growing
// rather than being drawn in full.
type svgWriter struct {
	b bytes.Buffer
}

func (w *svgWriter) full() bool { return w.b.Len() > MaxOutputLen }

// document wraps what w holds, a drawing bodyW wide, in an <svg> element of
// width by height labelled for assistive technology, under the title if
// there is one: the drawing is moved below the title and centered under it
// when the title is wider.
func (w *svgWriter) document(width, height, bodyW float64, title, label, desc string) []byte {
	dx, dy := (width-bodyW)/2, titleHeight(title)

	var b svgWriter
	b.b.Grow(w.b.Len() + 512)
	b.b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" class="gno-mermaid__svg"`)
	b.attr("width", width)
	b.attr("height", height)
	b.b.WriteString(` viewBox="0 0 `)
	writeNum(&b.b, width)
	b.b.WriteByte(' ')
	writeNum(&b.b, height)
	b.b.WriteString(`" font-size="`)
	writeNum(&b.b, fontSize)
	b.b.WriteString(`" role="img" aria-label="`)
	escape(&b.b, label)
	b.b.WriteString(`">`)
	if desc != "" {
		b.b.WriteString("<desc>")
		escape(&b.b, desc)
		b.b.WriteString("</desc>")
	}
	b.text(classTitle, width/2, pad+titleSize, anchorMiddle, titleSize, title)
	if dx != 0 || dy != 0 {
		b.b.WriteString(`<g transform="translate(`)
		writeNum(&b.b, dx)
		b.b.WriteByte(' ')
		writeNum(&b.b, dy)
		b.b.WriteString(`)">`)
		b.b.Write(w.b.Bytes())
		b.b.WriteString("</g>")
	} else {
		b.b.Write(w.b.Bytes())
	}
	b.b.WriteString("</svg>")
	return b.b.Bytes()
}

func (w *svgWriter) group(c class) {
	if w.open("g", c) {
		w.b.WriteByte('>')
	}
}

func (w *svgWriter) end() {
	if !w.full() {
		w.b.WriteString("</g>")
	}
}

func (w *svgWriter) rect(c class, x, y, width, height, radius float64) {
	if !w.open("rect", c) {
		return
	}
	w.attr("x", x)
	w.attr("y", y)
	w.attr("width", width)
	w.attr("height", height)
	if radius > 0 {
		w.attr("rx", radius)
	}
	w.b.WriteString("/>")
}

func (w *svgWriter) line(c class, x1, y1, x2, y2 float64) {
	if !w.open("line", c) {
		return
	}
	w.attr("x1", x1)
	w.attr("y1", y1)
	w.attr("x2", x2)
	w.attr("y2", y2)
	w.b.WriteString("/>")
}

func (w *svgWriter) circle(c class, cx, cy, r float64) {
	if !w.open("circle", c) {
		return
	}
	w.attr("cx", cx)
	w.attr("cy", cy)
	w.attr("r", r)
	w.b.WriteString("/>")
}

func (w *svgWriter) polygon(c class, xy ...float64) { w.points("polygon", c, xy) }

func (w *svgWriter) polyline(c class, xy ...float64) { w.points("polyline", c, xy) }

func (w *svgWriter) points(name string, c class, xy []float64) {
	if !w.open(name, c) {
		return
	}
	w.b.WriteString(` points="`)
	for i, v := range xy {
		switch {
		case i == 0:
		case i%2 == 1:
			w.b.WriteByte(',')
		default:
			w.b.WriteByte(' ')
		}
		writeNum(&w.b, v)
	}
	w.b.WriteString(`"/>`)
}

func (w *svgWriter) path(c class, d *pathData) {
	if !w.open("path", c) {
		return
	}
	w.b.WriteString(` d="`)
	w.b.Write(d.b.Bytes())
	w.b.WriteString(`"/>`)
}

// text writes lines of text, the first baseline at y and each next one
// leading ems lower. A size of 0 keeps the document's font size.
func (w *svgWriter) text(c class, x, y float64, a anchor, size float64, lines ...string) {
	if len(lines) == 0 || len(lines) == 1 && lines[0] == "" {
		return
	}
	if !w.open("text", c) {
		return
	}
	w.attr("x", x)
	w.attr("y", y)
	if a != anchorStart {
		w.b.WriteString(` text-anchor="`)
		w.b.WriteString(string(a))
		w.b.WriteByte('"')
	}
	if size == 0 {
		size = fontSize
	} else if size != fontSize {
		w.attr("font-size", size)
	}
	w.b.WriteByte('>')
	if len(lines) == 1 {
		escape(&w.b, lines[0])
	} else {
		for i, l := range lines {
			w.b.WriteString("<tspan")
			w.attr("x", x)
			w.attr("y", y+float64(i)*size*leading)
			w.b.WriteByte('>')
			escape(&w.b, l)
			w.b.WriteString("</tspan>")
		}
	}
	w.b.WriteString("</text>")
}

// open starts an element, or reports false when w is full.
func (w *svgWriter) open(name string, c class) bool {
	if w.full() {
		return false
	}
	w.b.WriteByte('<')
	w.b.WriteString(name)
	if c != "" {
		w.b.WriteString(` class="`)
		w.b.WriteString(string(c))
		w.b.WriteByte('"')
	}
	return true
}

func (w *svgWriter) attr(name string, v float64) {
	w.b.WriteByte(' ')
	w.b.WriteString(name)
	w.b.WriteString(`="`)
	writeNum(&w.b, v)
	w.b.WriteByte('"')
}

// pathData builds the d attribute of a path from numbers only.
type pathData struct {
	b bytes.Buffer
}

func (p *pathData) cmd(c byte, args ...float64) *pathData {
	p.b.WriteByte(c)
	for i, v := range args {
		if i > 0 {
			p.b.WriteByte(' ')
		}
		writeNum(&p.b, v)
	}
	return p
}

// maxCoord bounds the numbers writeNum writes, far beyond any drawing
// Render accepts (maxDimension), so a number is always a plain decimal.
const maxCoord = 1e9

// writeNum writes v with at most two decimals. A value that is not finite,
// which no layout should produce, is written as 0, and one past maxCoord as
// the bound, so the document stays valid.
func writeNum(b *bytes.Buffer, v float64) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		v = 0
	}
	v = math.Round(min(max(v, -maxCoord), maxCoord)*100) / 100
	if v == 0 {
		v = 0 // not -0
	}
	b.Write(strconv.AppendFloat(b.AvailableBuffer(), v, 'f', -1, 64))
}

// escape writes s as XML character data, valid in text and in attribute
// values: markup characters become references and characters XML forbids
// become U+FFFD.
func escape(b *bytes.Buffer, s string) {
	_ = xml.EscapeText(b, []byte(s))
}
