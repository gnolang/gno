package logos

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

//----------------------------------------
// Page

// A Page has renderable Elem(ents).
type Page struct {
	Size
	Style
	Attrs
	Elems  []Elem
	Cursor int // selected cursor element index, or -1.
}

// An elem is something that can draw a portion of itself onto a view.
// It has a relative coord and a size.
// Before it is drawn, it is rendered.
// Measure will update its size, while GetSize() returns the cached size.
type Elem interface {
	Render()
	Draw(offset Coord, dst View)
	GetCoord() Coord
	GetStyle() *Style
	GetAttrs() *Attrs
	SetIsCursor(bool)
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
			te.Coord = Coord{X: xpos, Y: ypos}
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
		Elems:  elems,
		Cursor: -1,
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
   Let's say we want to draw a Page.  We want to draw it onto
   some buffer, or more specially some "view" of a buffer (as
   a slice of an array is a view into an array buffer).

     Page virtual bounds:
     0 - - - - - - - - - - +
     :\                    :
     : \ (3,3)             :
     :  @-----------+      :
     :  |View       |      :
     :  |           |      :
     :  +-----------+      :
     :                     :
     + - - - - - - - - - - +

   0 is the origin point for the Page.
   @ is an offset within the page. @ here is (3,3).
   It is where the View is conceptually placed, but
   otherwise the View isn't aware where @ is.
   This offset is passed in as an argument 'offset'.

   NOTE: Offset is relative in the base page.  To offset the
   drawing position in the view (e.g. to only write on the
   right half of the buffer view), derive another view from
   the original.

   The View is associated with an underlying (base) buffer.

       Page virtual bounds:
       0 - - - - - - - - - - +
     +=:=====================:===+ <-- underlying Buffer
     | :                     :   |
     | :  @-------------+    :   |
     | :  |View         |    :   |
     | :  |.Offset=(5,2)|    :   |
     | :  +-------------+    :   |
     | :                     :   |
     | + - - - - - - - - - - +   |
     +===========================+

   Each element must be drawn onto the buffer view with the
   right offset algebra applied.  Here is a related diagram
   showing the buffer in relation to page elements.

      Page virtual bounds:
      0 - - - - - - - - - - - +
      :elem 1     |elem 2     :
      :       @------------+  :
      :       |View        |  :
      + - - - | - E - - - -|- +
      :elem 3 |   |elem 4  |  :
      :       +------------+  :
      :           |           :
      + - - - - - - - - - - - +

   In this example the page is composed of four element tiles.
   E is elem-4's offset relative to 0, the page's origin.  To
   draw the top-left portion of elem-4 onto the buffer slice
   as shown, the element is drawn with an offset of @-E, which
   is negative and indicates that the element should be drawn
   offset positively (right and bottom) from @.
*/

// Unlike TextElem or BufferedPageView, a Page doesn't keep
// its own buffer.  Its render function calls the elements'
// render functions, and the element buffers are combined
// during Draw(). There is a need for distinction because
// Draw() can't be too slow, so Render() is about optimizing
// Draw() calls.  The distinction between *Page and
// BufferedPageView gives the user more flexibility.
func (pg *Page) Render() {
	for _, elem := range pg.Elems {
		elem.Render()
	}
	if debug {
		debug.Println("sleeping after rendering page elements")
		time.Sleep(time.Second)
	}
}

// Draw the rendered page elements onto the view.
func (pg *Page) Draw(offset Coord, view View) {
	// First, draw page background style.
	style := pg.Style
	for x := 0; x < view.Bounds.Width; x++ {
		for y := 0; y < view.Bounds.Height; y++ {
			cell := view.GetCell(x, y)
			cell.Foreground = style.Foreground
			cell.Background = style.Foreground
			cell.StyleFlags = style.StyleFlags
		}
	}
	// Then, draw elems.
	for _, elem := range pg.Elems {
		eoffset := offset.Sub(elem.GetCoord())
		elem.Draw(eoffset, view)
	}
	if debug {
		debug.Println("sleeping after drawing page")
		time.Sleep(time.Second)
	}
}

type EventKey = tcell.EventKey

func (pg *Page) ProcessEventKey(ev *EventKey) {
	switch ev.Key() {
	case tcell.KeyEsc:
	case tcell.KeyUp:
	case tcell.KeyDown:
		pg.IncCursor()
	case tcell.KeyLeft:
	case tcell.KeyRight:
	case tcell.KeyEnter:
	default:
	}
}

func (pg *Page) IncCursor() {
	if pg.Cursor == -1 {
		if len(pg.Elems) == 0 {
			// nothing to select.
		} else {
			pg.Cursor = 0
			pg.Elems[pg.Cursor].SetIsCursor(true)
		}
	} else {
		pg.Elems[pg.Cursor].SetIsCursor(false)
		pg.Cursor++
		if pg.Cursor == len(pg.Elems) {
			pg.Cursor = 0 // roll back.
		}
		pg.Elems[pg.Cursor].SetIsCursor(true)
	}
}

//----------------------------------------
// TextElem

type TextElem struct {
	Coord
	Size
	Style
	Attrs
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
	style := tel.Style.WithAttrs(&tel.Attrs)
	runes := toRunes(tel.Text)
	i := 0
	for 0 < len(runes) {
		s, w, n := nextCharacter(runes)
		if n == 0 {
			panic(fmt.Sprintf(
				"unexpected error reading next character from runes %v",
				runes))
		} else {
			runes = runes[n:]
		}
		cell := tel.Buffer.GetCell(i, 0)
		cell.SetValue(s, w, style, tel)
		i += w
	}
	if i != tel.Buffer.Width {
		panic(fmt.Sprintf(
			"wrote %d cells but there are %d in buffer with text %q",
			i, tel.Buffer.Width, tel.Text))
	}
}

func (tel *TextElem) Draw(offset Coord, view View) {
	minX, maxX, minY, maxY := computeIntersection(tel.Size, offset, view.Bounds)
	for y := minY; y < maxY; y++ {
		if minY != 0 {
			panic("should not happen")
		}
		for x := minX; x < maxX; x++ {
			bcell := tel.Buffer.GetCell(x, y)
			vcell := view.GetCell(x-offset.X, y-offset.Y)
			vcell.SetCell(bcell)
		}
	}
}

type Color = tcell.Color

// Style is purely visual and has no side effects.
type Style struct {
	Foreground Color
	Background Color
	Padding    Padding
	StyleFlags
	Other []KVPair
}

func (st *Style) GetStyle() *Style {
	return st
}

func (stv Style) WithAttrs(attrs *Attrs) Style {
	if attrs.GetIsCursor() {
		stv.Background = tcell.ColorYellow
		return stv
	} else {
		return stv
	}
}

type StyleFlags uint32

const (
	StyleFlagNone StyleFlags = 0
	StyleFlagBold StyleFlags = 1 << iota
	StyleFlagBlink
	StyleFlagUnderline
	StyleFlagItalic
	StyleFlagStrikeThrough
)

// Attrs have side effects in the Logos system;
// for example, the lone cursor element (one with AttrFlagIsCursor set)
// is where most key events are sent to.
type Attrs struct {
	AttrFlags
	Other []KVPair
}

func (tt *Attrs) GetAttrs() *Attrs {
	return tt
}

func (tt *Attrs) GetIsCursor() bool {
	return (tt.AttrFlags & AttrFlagIsCursor) != 0
}

func (tt *Attrs) SetIsCursor(c bool) {
	if c {
		tt.AttrFlags |= AttrFlagIsCursor
	} else {
		tt.AttrFlags &= ^AttrFlagIsCursor
	}
}

type AttrFlags uint32

const (
	AttrFlagNone       AttrFlags = 0
	AttrFlagIsCursor   AttrFlags = 1 << iota // is current cursor
	AttrFlagIsSelected                       // is selected (among possibly others)
	AttrFlagIsDirty                          // is dirty (not yet used)
)

type KVPair struct {
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
		/*
			View
			+----------+
			|   [Elem__|____]
			+----------+       */
		minX = 0
	} else {
		/*
			     View
			     +----------+
			[____|__Elem]   |
			     +----------+  */
		minX = elo.X
	}
	if vws.Width < elo.X+els.Width {
		/*
			View
			+----------+
			|   [Elem__|____]
			+----------+       */
		maxX = vws.Width - elo.X
	} else {
		/*
			     View
			     +----------+
			[____|__Elem]   |
			     +----------+  */
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

func (sz Size) IsZero() bool {
	return sz.Width == 0 && sz.Height == 0
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
