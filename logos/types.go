package logos

import (
	"fmt"

	"github.com/gdamore/tcell"
)

//----------------------------------------
// Buffer & View

// A Buffer is a buffer area in which to draw.
type Buffer struct {
	Size
	Cells []Cell
}

func NewBuffer(sz Size) *Buffer {
	return &Buffer{
		Size:  sz,
		Cells: make([]Cell, sz.Width*sz.Height),
	}
}

// offset elements must be 0 or positive.
func (bb *Buffer) NewView(offset Coord) View {
	if !offset.IsNonNegative() {
		panic("should not happen")
	}
	return View{
		Base:   bb,
		Offset: offset,
		Size:   bb.Size,
	}
}

func (bb *Buffer) GetCell(x, y int) *Cell {
	if bb.Size.Width <= x {
		panic("should not happen")
	}
	if bb.Size.Height <= y {
		panic("should not happen")
	}
	return &bb.Cells[y*bb.Size.Width*x]
}

// Offset and Size must be within bounds of *Buffer.
// Here is an analogy: "Buffer:View :: array:slice".
type View struct {
	Base   *Buffer
	Offset Coord // offset within Buffer
	Size         // total size of slice
}

func (bs View) NewView(offset Coord) View {
	return View{
		Base:   bs.Base,
		Offset: bs.Offset.Add(offset),
		Size:   bs.Size.SubCoord(offset),
	}
}

func (bs View) GetCell(x, y int) *Cell {
	if bs.Size.Width <= x {
		panic("should not happen")
	}
	if bs.Size.Height <= y {
		panic("should not happen")
	}
	return bs.Base.GetCell(
		bs.Offset.X+x,
		bs.Offset.Y+y,
	)
}

//----------------------------------------
// Page

// A Page has renderable Elem(ents).
type Page struct {
	Size
	Style
	Elems []Elem
}

// An elem is something that can draw a portion of itself onto a view.
// It has a relative coord and a size.
// Before it is drawn, it is rendered.
// Measure will update its size, while GetSize() returns the cached size.
type Elem interface {
	Render()
	Draw(offset Coord, dst View)
	GetCoord() Coord
	GetSize() Size
	Measure() Size
}

// produces a page from a string.
// width is the width of the page.
// if isCode, width is ignored.
func NewPage(s string, width int, isCode bool, style Style) *Page {
	elems := []Elem{}
	ypos := 0
	xpos := 0
	lines := splitLines(s)
	if isCode {
		for _, line := range lines {
			te := NewTextElem(line, style)
			elems = append(elems, te)
			ypos++
			xpos = 0
		}
	} else {
		for _, line := range lines {
			words := splitSpaces(line)
			for _, word := range words {
				wd := widthOf(word)
				if width < xpos+wd {
					if xpos != 0 {
						ypos++
						xpos = 0
					}
				}
				te := NewTextElem(word, style)
				te.Coord = Coord{X: xpos, Y: ypos}
				elems = append(elems, te)
				xpos += te.Width // size of word
				xpos += 1        // space after each word (not written)
			}
		}
	}
	page := &Page{
		Size: Size{
			Width:  width,
			Height: -1, // not set
		},
		Elems: elems,
	}
	page.Measure()
	return page
}

// Assumes page starts at 0,0.
func (pg *Page) Measure() Size {
	maxX := 0
	maxY := 0
	for _, view := range pg.Elems {
		coord := view.GetCoord()
		size := view.GetSize()
		if maxX < coord.X+size.Width {
			maxX = coord.X + size.Width
		}
		if maxY < coord.Y+size.Height {
			maxY = coord.Y + size.Height
		}
	}
	size := Size{
		Width:  maxX,
		Height: maxY,
	}
	pg.Size = size
	return size
}

/*
   Page draw logic:
   Let's say we want to draw a Page.  We want to draw it onto some buffer,
   or more specially some "view" of a buffer (as a slice of an array is
   a view into an array buffer).

     Page virtual bounds:
     0 - - - - - - - - - - +
     :\                    :
     : \ (3,3)             :
     :  P-----------+      :
     :  |View       |      :
     :  |           |      :
     :  +-----------+      :
     :                     :
     + - - - - - - - - - - +

   0 is the origin point for the Page.
   P is an offset within the page.  P here is (3,3).
   This offset is passed in as an argument 'offset'.

   NOTE: Offset is relative in the base page.  To offset the drawing position
   in the view (e.g. to only write on the right half of the buffer view),
   derive another view from the original.

   The View is associated with an underlying (base) buffer.

       Page virtual bounds:
       0 - - - - - - - - - - +
     +=:=====================:===+ <-- underlying Buffer
     | :                     :   |
     | :  P-------------+    :   |
     | :  |View         |    :   |
     | :  |.Offset=(5,2)|    :   |
     | :  +-------------+    :   |
     | :                     :   |
     | + - - - - - - - - - - +   |
     +===========================+

   Each element must be drawn onto the buffer view with the right offset
   algebra applied.  Here is a related diagram showing the buffer in relation
   to page elements.

      Page virtual bounds:
      0 - - - - - - - - - - - +
      :elem 1     |elem 2     :
      :       P------------+  :
      :       |View        |  :
      + - - - | - E - - - -|- +
      :elem 3 |   |elem 4  |  :
      :       +------------+  :
      :           |           :
      + - - - - - - - - - - - +

   In this example the page is composed of four element tiles.  E is elem-4's
   offset relative to 0, the page's origin.  To draw the top-left portion of
   elem-4 onto the buffer slice as shown, the element is drawn with an offset
   of P-E, which is negative and indicates that the element should be drawn
   offset positively (right and bottom) from P.
*/

// Unlike TextElem or BufferedPageView, a Page doesn't keep its own buffer.
// Its render function calls the elements' render functions, and the element
// buffers are combined during Draw(). There is a need for distinction because
// Draw() can't be too slow, so Render() is about optimizing Draw() calls.
// The distinction between *Page and BufferedPageView gives the user more flexibility.
func (pg *Page) Render() {
	for _, elem := range pg.Elems {
		elem.Render()
	}
}

// Draw the rendered page elements onto the view.
func (pg *Page) Draw(offset Coord, view View) {
	// First, draw page background style.
	style := pg.Style
	for x := 0; x < view.Size.Width; x++ {
		for y := 0; y < view.Size.Height; y++ {
			cell := view.GetCell(x, y)
			cell.Foreground = style.Foreground
			cell.Background = style.Foreground
			cell.Flags = style.Flags
		}
	}
	// Then, draw elems.
	for _, elem := range pg.Elems {
		eoffset := offset.Sub(elem.GetCoord())
		elem.Draw(eoffset, view)
	}
}

type TextElem struct {
	Coord
	Size
	Style
	Text string
	*Buffer
}

func NewTextElem(text string, style Style) *TextElem {
	te := &TextElem{
		Style: style,
		Text:  text,
		Buffer: NewBuffer(Size{
			Height: 1,
			Width:  widthOf(text),
		}),
	}
	te.Measure()
	return te
}

func (tel *TextElem) Measure() Size {
	size := Size{
		Height: 1,
		Width:  widthOf(tel.Text),
	}
	tel.Size = size
	return size
}

func (tel *TextElem) Render() {
	if tel.Height != 1 {
		panic("should not happen")
	}
	runes := toRunes(tel.Text)
	x, y := tel.X, tel.Y
	i := 0
	for {
		s, w, n := nextCharacter(runes)
		if n == 0 {
			panic("should not happen")
		}
		cell := tel.Buffer.GetCell(x+i, y)
		cell.SetValue(s, w, tel.Style, tel)
		if 1 < w {
			for j := 1; j < w; j++ {
				i++
				// TODO: do we need to set background color etc?
				// cell := tel.Buffer.GetCell(x+i, y)
				// cell.SetCharacter(" ", 1)
			}
			continue
		} else {
			i++
		}
	}
	if i != tel.Buffer.Width {
		panic(fmt.Sprintf(
			"wrote %d cells but there are %d in buffer",
			i, tel.Buffer.Width))
	}
}

func (tel *TextElem) Draw(offset Coord, view View) {
	minX, maxX, minY, maxY := computeIntersection(tel.Size, offset, view.Size)
	for y := minY; y < maxY; y++ {
		if minY != 0 {
			panic("should not happen")
		}
		for x := minX; x < maxX; x++ {
			bcell := tel.Buffer.GetCell(x, y)
			vcell := view.GetCell(x-minX, y-minY)
			vcell.SetCell(bcell)
		}
	}
}

type Color = tcell.Color

type Style struct {
	Foreground Color
	Background Color
	Padding    Padding
	Flags      Flags
	Other      []KVAttr
}

type Flags uint32

const (
	FlagNone Flags = 0
	FlagBold Flags = 1 << iota
	FlagBlink
	FlagUnderline
	FlagItalic
	FlagStrikeThrough
)

type KVAttr struct {
	Key   string
	Value interface{}
}

//----------------------------------------
// computeIntersection()

// els: element size
// elo: offset within element
// vws: view size
// minX,maxX,minY,maxY are relative to el.
// maxX and maxY are exclusive.
func computeIntersection(els Size, elo Coord, vws Size) (minX, maxX, minY, maxY int) {
	if elo.X < 0 {
		/*	View
			+----------+
			|   [Elem__|____]
			+----------+       */
		minX = 0
	} else {
		/* 	 View
			 +----------+
		[____|__Elem]   |
			 +----------+      */
		minX = elo.X
	}
	if vws.Width < elo.X+els.Width {
		/*	View
			+----------+
			|   [Elem__|____]
			+----------+       */
		maxX = vws.Width - elo.X
	} else {
		/*  View
			 +----------+
		[____|__Elem]   |
			 +----------+      */
		maxX = els.Width
	}
	if elo.Y < 0 {
		minY = 0
	} else {
		minY = elo.Y
	}
	if vws.Height < elo.Y+els.Height {
		maxY = vws.Height - elo.Y
	} else {
		maxY = els.Height
	}
	return
}

//----------------------------------------
// Misc simple types

type Padding struct {
	Top    int
	Left   int
	Right  int
	Bottom int
}

func (pd Padding) GetPadding() Padding {
	return pd
}

type Size struct {
	Width  int
	Height int // -1 if not set.
}

func (sz Size) GetSize() Size {
	return sz
}

// zero widths or heights are valid.
func (sz Size) IsValid() bool {
	return 0 <= sz.Width && 0 <= sz.Height
}

func (sz Size) IsPositive() bool {
	return 0 < sz.Width && 0 < sz.Height
}

func (sz Size) SubCoord(crd Coord) Size {
	if !crd.IsNonNegative() {
		panic("should not happen")
	}
	sz2 := Size{
		Width:  sz.Width - crd.X,
		Height: sz.Height - crd.Y,
	}
	if !sz2.IsValid() {
		panic("should not happen")
	}
	return sz2
}

type Coord struct {
	X int
	Y int
}

func (crd Coord) GetCoord() Coord {
	return crd
}

func (crd Coord) IsNonNegative() bool {
	return 0 <= crd.X && 0 <= crd.Y
}

func (crd Coord) Add(crd2 Coord) Coord {
	return Coord{
		X: crd.X + crd2.X,
		Y: crd.Y + crd2.Y,
	}
}

func (crd Coord) Sub(crd2 Coord) Coord {
	return Coord{
		X: crd.X - crd2.X,
		Y: crd.Y - crd2.Y,
	}
}
