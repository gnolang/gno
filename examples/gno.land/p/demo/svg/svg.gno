package svg

import (
	"encoding/base64"

	"gno.land/p/demo/ufmt"
)

type Canvas struct {
	Width, Height int
	ViewBox       string
	Elems         []Elem
}

type Elem interface{ String() string }

func (c Canvas) String() string {
	output := ""
	output += ufmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg" viewBox="%s">`, c.Width, c.Height, c.ViewBox)
	for _, elem := range c.Elems {
		output += elem.String()
	}
	output += "</svg>"
	return output
}

func (c Canvas) Base64() string {
	output := c.String()
	return base64.StdEncoding.EncodeToString([]byte(output))
}

func (c *Canvas) Append(elem ...Elem) {
	c.Elems = append(c.Elems, elem...)
}

type BaseAttrs struct {
	ID          string
	Class       string
	Style       string
	Fill        string
	Stroke      string
	StrokeWidth string
	Opacity     string
	Transform   string
	Visibility  string
}

func (b BaseAttrs) String() string {
	return ufmt.Sprintf(` id="%s" class="%s" style="%s" fill="%s" stroke="%s" stroke-width="%s" opacity="%s" transform="%s" visibility="%s"`, b.ID, b.Class, b.Style, b.Fill, b.Stroke, b.StrokeWidth, b.Opacity, b.Transform, b.Visibility)
}

type Circle struct {
	CX   int // center X
	CY   int // center Y
	R    int // radius
	Fill string
	BaseAttrs
}

func (c Circle) String() string {
	return ufmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" %s />`, c.CX, c.CY, c.R, c.Fill, c.BaseAttrs.String())
}

func NewCircle(cx, cy, r int, fill string) *Circle {
	return &Circle{
		CX:   cx,
		CY:   cy,
		R:    r,
		Fill: fill,
	}
}

type Ellipse struct {
	CX   int // center X
	CY   int // center Y
	RX   int // radius X
	RY   int // radius Y
	Fill string
	BaseAttrs
}

func (e Ellipse) String() string {
	return ufmt.Sprintf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" %s />`, e.CX, e.CY, e.RX, e.RY, e.Fill, e.BaseAttrs.String())
}

func NewEllipse(cx, cy int, fill string) *Ellipse {
	return &Ellipse{
		CX:   cx,
		CY:   cy,
		Fill: fill,
	}
}

type Rectangle struct {
	X, Y, Width, Height int
	RX, RY              int // corner radiuses
	Fill                string
	BaseAttrs
}

func (r Rectangle) String() string {
	return ufmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" rx="%d" ry="%d" fill="%s" />`, r.X, r.Y, r.Width, r.Height, r.RX, r.RY, r.Fill)
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

type Path struct {
	D    string
	Fill string
	BaseAttrs
}

func (p Path) String() string {
	return ufmt.Sprintf(`<path d="%s" fill="%s" />`, p.D, p.Fill)
}

func NewPath(d, fill string) *Path {
	return &Path{
		D:    d,
		Fill: fill,
	}
}

type Polygon struct { // closed shape
	Points string
	Fill   string
	BaseAttrs
}

func (p Polygon) String() string {
	return ufmt.Sprintf(`<polygon points="%s" fill="%s" />`, p.Points, p.Fill)
}

func NewPolygon(points, fill string) *Polygon {
	return &Polygon{
		Points: points,
		Fill:   fill,
	}
}

type Polyline struct { // polygon but not necessarily closed
	Points string
	Fill   string
	BaseAttrs
}

func (p Polyline) String() string {
	return ufmt.Sprintf(`<polyline points="%s" fill="%s" />`, p.Points, p.Fill)
}

func NewPolyline(points, fill string) *Polyline {
	return &Polyline{
		Points: points,
		Fill:   fill,
	}
}

type Text struct {
	X, Y       int
	DX, DY     int // shift text pos horizontally/ vertically
	Rotate     string
	Text, Fill string
	BaseAttrs
}

func (c Text) String() string {
	return ufmt.Sprintf(`<text x="%d" y="%d" dx="%d" dy="%d" rotate="%s" fill="%s">%s</text>`, c.X, c.Y, c.DX, c.DY, c.Rotate, c.Fill, c.Text)
}

func NewText(x, y int, text, fill string) *Text {
	return &Text{
		X:    x,
		Y:    y,
		Text: text,
		Fill: fill,
	}
}

type Group struct {
	Elems []Elem
	Fill  string
	BaseAttrs
}

func (g Group) String() string {
	out := ""
	for _, e := range g.Elems {
		out += e.String()
	}
	return ufmt.Sprintf(`<g fill="%s">%s</g>`, g.Fill, out)
}
