package svg

import "gno.land/p/demo/ufmt"

type Canvas struct {
	Width  int
	Height int
	Elems  []Elem
}

type Elem interface{ String() string }

func (c Canvas) String() string {
	output := ""
	output += ufmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, c.Width, c.Height)
	for _, elem := range c.Elems {
		output += elem.String()
	}
	output += "</svg>"
	return output
}

func (c *Canvas) Append(elem Elem) {
	c.Elems = append(c.Elems, elem)
}

type Circle struct {
	CX   int // center X
	CY   int // center Y
	R    int // radius
	Fill string
}

func (c Circle) String() string {
	return ufmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="%s" />`, c.CX, c.CY, c.R, c.Fill)
}

func (c *Canvas) DrawCircle(cx, cy, r int, fill string) {
	c.Append(Circle{
		CX:   cx,
		CY:   cy,
		R:    r,
		Fill: fill,
	})
}

type Rectangle struct {
	X, Y, Width, Height int
	Fill                string
}

func (c Rectangle) String() string {
	return ufmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="%s" />`, c.X, c.Y, c.Width, c.Height, c.Fill)
}

func (c *Canvas) DrawRectangle(x, y, width, height int, fill string) {
	c.Append(Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Fill:   fill,
	})
}

type Text struct {
	X, Y       int
	Text, Fill string
}

func (c Text) String() string {
	return ufmt.Sprintf(`<text x="%d" y="%d" fill="%s">%s</text>`, c.X, c.Y, c.Fill, c.Text)
}

func (c *Canvas) DrawText(x, y int, text, fill string) {
	c.Append(Text{
		X:    x,
		Y:    y,
		Text: text,
		Fill: fill,
	})
}
