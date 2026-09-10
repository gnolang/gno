package svg

import (
	"encoding/base64"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

type Canvas struct {
	Width, Height int
	ViewBox       string
	Elems         []Elem
	Style         *avl.Tree
}

type Elem interface{ String() string }

func NewCanvas(width, height int) *Canvas {
	return &Canvas{
		Width:  width,
		Height: height,
		Style:  nil,
	}
}

func (c *Canvas) AddStyle(key, value string) *Canvas {
	if c.Style == nil {
		c.Style = avl.NewTree()
	}
	c.Style.Set(key, value)
	return c
}

func (c *Canvas) WithViewBox(x, y, width, height int) *Canvas {
	c.ViewBox = ufmt.Sprintf("%d %d %d %d", x, y, width, height)
	return c
}

// Render renders your canvas
func (c Canvas) Render(alt string) string {
	base64SVG := base64.StdEncoding.EncodeToString([]byte(c.String()))
	return ufmt.Sprintf("![%s](data:image/svg+xml;base64,%s)", alt, base64SVG)
}

func (c Canvas) String() string {
	out := ""
	out += ufmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="%s">`, c.Width, c.Height, c.ViewBox)
	if c.Style != nil {
		out += "<style>"
		c.Style.Iterate("", "", func(k string, val interface{}) bool {
			v := val.(string)
			out += ufmt.Sprintf("%s{%s}", k, v)
			return false
		})
		out += "</style>"
	}
	for _, elem := range c.Elems {
		out += elem.String()
	}
	out += "</svg>"
	return out
}

func (c Canvas) Base64() string {
	out := c.String()
	return base64.StdEncoding.EncodeToString([]byte(out))
}

func (c *Canvas) Append(elem ...Elem) {
	c.Elems = append(c.Elems, elem...)
}

type BaseAttrs struct {
	ID          string
	Class       string
	Style       string
	Stroke      string
	StrokeWidth string
	Opacity     string
	Transform   string
	Visibility  string
}

func (b BaseAttrs) String() string {
	var elems []string

	if b.ID != "" {
		elems = append(elems, `id="`+b.ID+`"`)
	}
	if b.Class != "" {
		elems = append(elems, `class="`+b.Class+`"`)
	}
	if b.Style != "" {
		elems = append(elems, `style="`+b.Style+`"`)
	}
	if b.Stroke != "" {
		elems = append(elems, `stroke="`+b.Stroke+`"`)
	}
	if b.StrokeWidth != "" {
		elems = append(elems, `stroke-width="`+b.StrokeWidth+`"`)
	}
	if b.Opacity != "" {
		elems = append(elems, `opacity="`+b.Opacity+`"`)
	}
	if b.Transform != "" {
		elems = append(elems, `transform="`+b.Transform+`"`)
	}
	if b.Visibility != "" {
		elems = append(elems, `visibility="`+b.Visibility+`"`)
	}
	if len(elems) == 0 {
		return ""
	}
	return strings.Join(elems, " ")
}

type Circle struct {
	CX   int // center X
	CY   int // center Y
	R    int // radius
	Fill string
	Attr BaseAttrs
}

func (c Circle) String() string {
	return ufmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" %s/>`, c.CX, c.CY, c.R, c.Fill, c.Attr.String())
}

func NewCircle(cx, cy, r int, fill string) *Circle {
	return &Circle{
		CX:   cx,
		CY:   cy,
		R:    r,
		Fill: fill,
	}
}

func (c *Circle) WithClass(class string) *Circle {
	c.Attr.Class = class
	return c
}

type Ellipse struct {
	CX   int // center X
	CY   int // center Y
	RX   int // radius X
	RY   int // radius Y
	Fill string
	Attr BaseAttrs
}

func (e Ellipse) String() string {
	return ufmt.Sprintf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" %s/>`, e.CX, e.CY, e.RX, e.RY, e.Fill, e.Attr.String())
}

func NewEllipse(cx, cy int, fill string) *Ellipse {
	return &Ellipse{
		CX:   cx,
		CY:   cy,
		Fill: fill,
	}
}

func (e *Ellipse) WithClass(class string) *Ellipse {
	e.Attr.Class = class
	return e
}

type Rectangle struct {
	X, Y, Width, Height int
	RX, RY              int // corner radiuses
	Fill                string
	Attr                BaseAttrs
}

func (r Rectangle) String() string {
	return ufmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%d" ry="%d" fill="%s" %s/>`, r.X, r.Y, r.Width, r.Height, r.RX, r.RY, r.Fill, r.Attr.String())
}

func NewRectangle(x, y, width, height int, fill string) *Rectangle {
	return &Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Fill:   fill,
	}
}

func (r *Rectangle) WithClass(class string) *Rectangle {
	r.Attr.Class = class
	return r
}

type Path struct {
	D    string
	Fill string
	Attr BaseAttrs
}

func (p Path) String() string {
	return ufmt.Sprintf(`<path d="%s" fill="%s" %s/>`, p.D, p.Fill, p.Attr.String())
}

func NewPath(d, fill string) *Path {
	return &Path{
		D:    d,
		Fill: fill,
	}
}

func (p *Path) WithClass(class string) *Path {
	p.Attr.Class = class
	return p
}

type Polygon struct { // closed shape
	Points string
	Fill   string
	Attr   BaseAttrs
}

func (p Polygon) String() string {
	return ufmt.Sprintf(`<polygon points="%s" fill="%s" %s/>`, p.Points, p.Fill, p.Attr.String())
}

func NewPolygon(points, fill string) *Polygon {
	return &Polygon{
		Points: points,
		Fill:   fill,
	}
}

func (p *Polygon) WithClass(class string) *Polygon {
	p.Attr.Class = class
	return p
}

type Polyline struct { // polygon but not necessarily closed
	Points string
	Fill   string
	Attr   BaseAttrs
}

func (p Polyline) String() string {
	return ufmt.Sprintf(`<polyline points="%s" fill="%s" %s/>`, p.Points, p.Fill, p.Attr.String())
}

func NewPolyline(points, fill string) *Polyline {
	return &Polyline{
		Points: points,
		Fill:   fill,
	}
}

func (p *Polyline) WithClass(class string) *Polyline {
	p.Attr.Class = class
	return p
}

type Text struct {
	X, Y       int
	DX, DY     int // shift text pos horizontally/ vertically
	Rotate     string
	Text, Fill string
	Attr       BaseAttrs
}

func (c Text) String() string {
	return ufmt.Sprintf(`<text x="%d" y="%d" dx="%d" dy="%d" rotate="%s" fill="%s" %s>%s</text>`, c.X, c.Y, c.DX, c.DY, c.Rotate, c.Fill, c.Attr.String(), c.Text)
}

func NewText(x, y int, text, fill string) *Text {
	return &Text{
		X:    x,
		Y:    y,
		Text: text,
		Fill: fill,
	}
}

func (c *Text) WithClass(class string) *Text {
	c.Attr.Class = class
	return c
}

type Group struct {
	Elems []Elem
	Fill  string
	Attr  BaseAttrs
}

func (g Group) String() string {
	out := ""
	for _, e := range g.Elems {
		out += e.String()
	}
	return ufmt.Sprintf(`<g fill="%s" %s>%s</g>`, g.Fill, g.Attr.String(), out)
}

func NewGroup(fill string) *Group {
	return &Group{
		Fill: fill,
	}
}

func (g *Group) Append(elem ...Elem) {
	g.Elems = append(g.Elems, elem...)
}

func (g *Group) WithClass(class string) *Group {
	g.Attr.Class = class
	return g
}
