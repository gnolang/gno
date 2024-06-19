package logos

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ----------------------------------------
// Page

// A Page has renderable Elem(ents).
type Page struct {
	Coord // used by parent only. TODO document.
	Size
	*Style
	Attrs
	Elems  []Elem
	Cursor int // selected cursor element index, or -1.
}

// An elem is something that can draw a portion of itself onto
// a view.  It has a relative coord and a size.  Before it is
// drawn, it is rendered.  Measure will update its size, while
// GetSize() returns the cached size.  ProcessEventKey()
// returns true if event was consumed.
type Elem interface {
	GetParent() Elem
	SetParent(Elem)
	GetCoord() Coord
	SetCoord(Coord)
	GetStyle() *Style
	SetStyle(*Style)
	GetAttrs() *Attrs
	GetIsCursor() bool
	SetIsCursor(bool)
	GetIsDirty() bool
	SetIsDirty(bool)
	GetIsOccluded() bool
	SetIsOccluded(bool)
	GetSize() Size
	Measure() Size
	Render() bool
	Draw(offset Coord, dst View)
	ProcessEventKey(*EventKey) bool
	String() string
	StringIndented(indent string) string
	// NOTE: SetSize(Size) isn't an elem interface, as
	// containers in general can't force elements to be of a
	// certain size, but rather prefers drawing out of
	// bounds; this opinion may distinguishes Logos from
	// other most gui frameworks.
}

var (
	_ Elem = &Page{}
	_ Elem = &BufferedElemView{}
	_ Elem = &TextElem{}
	_ Elem = &Stack{}
)

// produces a page from a string.
// width is the width of the page.
// if isCode, width is ignored.
func NewPage(s string, width int, isCode bool, style *Style) *Page {
	page := &Page{
		Size: Size{
			Width:  width,
			Height: -1, // not set
		},
		Style:  style,
		Elems:  nil, // will set
		Cursor: -1,
	}
	elems := []Elem{}
	if s != "" {
		pad := style.GetPadding()
		ypos := 0 + pad.Top
		xpos := 0 + pad.Left
		lines := splitLines(s)
		if isCode {
			for _, line := range lines {
				te := NewTextElem(line, style)
				te.SetParent(page)
				te.SetCoord(Coord{X: xpos, Y: ypos})
				elems = append(elems, te)
				ypos++
				xpos = 0 + pad.Left
			}
		} else {
			for _, line := range lines {
				words := splitSpaces(line)
				for _, word := range words {
					wd := widthOf(word)
					if width < xpos+wd+pad.Left+pad.Right {
						if xpos != 0+pad.Left {
							ypos++
							xpos = 0 + pad.Left
						}
					}
					te := NewTextElem(word, style)
					te.SetParent(page)
					te.SetCoord(Coord{X: xpos, Y: ypos})
					elems = append(elems, te)
					xpos += te.Width // size of word
					xpos += 1        // space after each word (not written)
				}
				ypos++
				xpos = 0 + pad.Left
			}
		}
	}
	page.Elems = elems
	page.Measure()
	page.SetIsDirty(true)
	return page
}

func (pg *Page) StringIndented(indent string) string {
	elines := []string{}
	eindent := indent + "    "
	for _, elem := range pg.Elems {
		elines = append(elines, eindent+elem.StringIndented(eindent))
	}
	return fmt.Sprintf("Page%v@%p\n%s",
		pg.Size,
		pg,
		strings.Join(elines, "\n"))
}

func (pg *Page) String() string {
	return fmt.Sprintf("Page%v{%d}@%p",
		pg.Size, len(pg.Elems), pg)
}

func (pg *Page) NextCoord() Coord {
	if len(pg.Elems) == 0 {
		return Coord{X: pg.GetPadding().Left, Y: pg.GetPadding().Top}
	} else {
		last := pg.Elems[len(pg.Elems)-1]
		last.Measure()
		lcoord := last.GetCoord()
		lsize := last.GetSize()
		return Coord{
			X: pg.GetPadding().Left,
			Y: lcoord.Y + lsize.Height, // no spacers by spec.
		}
	}
}

func (pg *Page) SetStyle(style *Style) {
	pg.Style = style
}

// Measures the size of elem and appends to page below the last element,
// or if empty, the top-leftmost coordinate exclusive of padding cells.
func (pg *Page) AppendElem(elem Elem) {
	ncoord := pg.NextCoord()
	elem.SetParent(pg)
	elem.SetCoord(ncoord)
	pg.Elems = append(pg.Elems, elem)
	pg.SetIsDirty(true)
}

// Assumes page starts at 0,0.
func (pg *Page) Measure() Size {
	pad := pg.GetPadding()
	maxX := pad.Left
	maxY := pad.Top
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
		Width:  maxX + pad.Right,
		Height: maxY + pad.Bottom,
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

// Unlike TextElem or BufferedElemView, a Page doesn't keep
// its own buffer.  Its render function calls the elements'
// render functions, and the element buffers are combined
// during Draw(). There is a need for distinction because
// Draw() can't be too slow, so Render() is about optimizing
// Draw() calls.  The distinction between *Page and
// BufferedElemView gives the user more flexibility.
func (pg *Page) Render() (updated bool) {
	if !pg.GetIsDirty() {
		return
	} else {
		defer pg.SetIsDirty(false)
	}
	for _, elem := range pg.Elems {
		elem.Render()
	}
	return true
}

// Draw the rendered page elements onto the view.
func (pg *Page) Draw(offset Coord, view View) {
	style := pg.GetStyle()
	border := style.GetBorder()
	minX, maxX, minY, maxY := computeIntersection(pg.Size, offset, view.Bounds)
	// First, draw page background style.
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			xo, yo := x-offset.X, y-offset.Y
			vcell := view.GetCell(xo, yo)
			// Draw area and border.
			if x == 0 {
				if y == pg.Size.Height-1 {
					// handle this case first so if height is 1,
					// this corner is preferred.
					vcell.SetValue(border.BLCorner(), 1, style, pg)
				} else if y == 0 {
					vcell.SetValue(border.TLCorner(), 1, style, pg)
				} else {
					vcell.SetValue(border.LeftBorder(y), 1, style, pg)
				}
			} else if x == pg.Size.Width-1 {
				if y == pg.Size.Height-1 {
					// ditto for future left-right language support.
					vcell.SetValue(border.BRCorner(), 1, style, pg)
				} else if y == 0 {
					vcell.SetValue(border.TRCorner(), 1, style, pg)
				} else {
					vcell.SetValue(border.RightBorder(y), 1, style, pg)
				}
			} else if y == 0 {
				vcell.SetValue(border.TopBorder(x), 1, style, pg)
			} else if y == pg.Size.Height-1 {
				vcell.SetValue(border.BottomBorder(x), 1, style, pg)
			} else { // Draw area.
				vcell.SetValue(" ", 1, style, pg)
			}
		}
	}
	// Then, draw elems.
	for _, elem := range pg.Elems {
		eoffset := offset.Sub(elem.GetCoord())
		elem.Draw(eoffset, view)
	}
}

type EventKey = tcell.EventKey

func (pg *Page) ProcessEventKey(ev *EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEsc:
		return false
	case tcell.KeyUp:
		pg.DecCursor(true)
	case tcell.KeyDown:
		pg.IncCursor(true)
	case tcell.KeyLeft:
		pg.DecCursor(false)
	case tcell.KeyRight:
		pg.IncCursor(false)
	case tcell.KeyEnter:
		if pg.Cursor == -1 {
			// as if pressed down
			pg.IncCursor(true)
			return true
		}
		// XXX this is a test.
		st := StackOf(pg)
		celem := pg.Elems[pg.Cursor]
		coord := AbsCoord(celem).Sub(AbsCoord(st))
		page := NewPage("this is a test", 80, false, pg.Style)
		coord.Y += 1
		coord.X += 2
		page.SetCoord(coord)
		st.PushLayer(page)
	default:
		return false
	}
	// Leave as true for convenience in cases above.
	// If a key event wasn't consumed, return false.
	return true
}

func (pg *Page) IncCursor(isVertical bool) {
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

func (pg *Page) DecCursor(isVertical bool) {
	if pg.Cursor == -1 {
		if len(pg.Elems) == 0 {
			// nothing to select.
		} else {
			pg.Cursor = len(pg.Elems) - 1
			pg.Elems[pg.Cursor].SetIsCursor(true)
		}
	} else {
		pg.Elems[pg.Cursor].SetIsCursor(false)
		pg.Cursor--
		if pg.Cursor == -1 {
			pg.Cursor = len(pg.Elems) - 1 // roll forward.
		}
		pg.Elems[pg.Cursor].SetIsCursor(true)
	}
}

// ----------------------------------------
// TextElem

type TextElem struct {
	Coord
	Size
	*Style // ignores padding.
	Attrs
	Text string
	*Buffer
}

func NewTextElem(text string, style *Style) *TextElem {
	te := &TextElem{
		Style: style,
		Text:  text,
		Buffer: NewBuffer(Size{
			Height: 1,
			Width:  widthOf(text),
		}),
	}
	te.Measure()
	te.SetIsDirty(true)
	return te
}

func (tel *TextElem) SetStyle(style *Style) {
	tel.Style = style
}

func (tel *TextElem) StringIndented(indent string) string {
	return tel.String()
}

func (tel *TextElem) String() string {
	return fmt.Sprintf("Text{%q}", tel.Text)
}

func (tel *TextElem) Measure() Size {
	size := Size{
		Height: 1,
		Width:  widthOf(tel.Text),
	}
	tel.Size = size
	return size
}

func (tel *TextElem) Render() (updated bool) {
	if tel.Height != 1 {
		panic("should not happen")
	}
	if !tel.GetIsDirty() {
		return
	} else {
		defer tel.SetIsDirty(false)
	}
	tel.Buffer.Reset()
	style := tel.GetStyle()
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
		for j := 1; j < w; j++ {
			cell := tel.Buffer.GetCell(i+j, 0)
			cell.SetValue("", 0, style, tel) // clear next cells
		}
		i += w
	}
	if i != tel.Buffer.Width {
		panic(fmt.Sprintf(
			"wrote %d cells but there are %d in buffer with text %q",
			i, tel.Buffer.Width, tel.Text))
	}
	return true
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
			vcell.SetValueFromCell(bcell)
		}
	}
}

func (tel *TextElem) ProcessEventKey(ev *EventKey) bool {
	return false // TODO: clipboard.
}

// ----------------------------------------
// misc.

type Color = tcell.Color

// Style is purely visual and has no side effects.
// It is generally referred to by pointer; you may need to copy before
// modifying.
type Style struct {
	Foreground Color
	Background Color
	Padding    Padding
	Border     Border
	StyleFlags
	Other       []KVPair
	CursorStyle *Style
}

func DefaultStyle() *Style {
	return &Style{
		Foreground: gDefaultForeground,
		Background: gDefaultBackground,
		CursorStyle: &Style{
			Background: tcell.ColorYellow,
		},
	}
}

var (
	gDefaultStyle      = DefaultStyle()
	gDefaultForeground = tcell.ColorBlack
	gDefaultBackground = tcell.ColorLightBlue
)

func (st *Style) Copy() *Style {
	st2 := *st
	return &st2
}

func (st *Style) GetStyle() *Style {
	return st
}

func (st *Style) GetForeground() Color {
	if st == nil {
		return gDefaultStyle.Foreground
	} else {
		return st.Foreground
	}
}

func (st *Style) GetBackground() Color {
	if st == nil {
		return gDefaultStyle.Background
	} else {
		return st.Background
	}
}

func (st *Style) GetPadding() Padding {
	if st == nil {
		return gDefaultStyle.Padding
	} else {
		return st.Padding
	}
}

func (st *Style) GetBorder() *Border {
	if st == nil {
		return &gDefaultStyle.Border
	} else {
		return &st.Border
	}
}

func (st *Style) GetCursorStyle() *Style {
	if st == nil {
		return gDefaultStyle.CursorStyle
	} else if st.CursorStyle == nil {
		return st
	} else {
		return st.CursorStyle
	}
}

// NOTE: this should only be called during the last step when
// writing to screen.  The receiver must not be nil and must
// not be modified, and the result is a value, not the style
// of any particular element.
func (st *Style) WithAttrs(attrs *Attrs) (res Style) {
	if st == nil {
		panic("unexpected nil style")
	}
	if attrs.GetIsCursor() {
		res = *st.GetCursorStyle()
	} else {
		res = *st
	}
	if attrs.GetIsOccluded() {
		res.SetIsShaded(true)
	}
	return
}

func (st Style) GetTStyle() (tst tcell.Style) {
	if st.Foreground.Valid() {
		tst = tst.Foreground(st.Foreground)
	} else {
		tst = tst.Foreground(gDefaultForeground)
	}
	if st.Background.Valid() {
		tst = tst.Background(st.Background)
	} else {
		tst = tst.Background(gDefaultBackground)
	}
	if st.GetIsShaded() {
		tst = tst.Dim(true)
		tst = tst.Background(tcell.ColorGray)
	}
	// TODO StyleFlags
	return tst
}

type StyleFlags uint32

func (sf StyleFlags) GetIsDim() bool {
	return (sf & StyleFlagDim) != 0
}

func (sf *StyleFlags) SetIsDim(id bool) {
	if id {
		*sf |= StyleFlagDim
	} else {
		*sf &= ^StyleFlagDim
	}
}

func (sf StyleFlags) GetIsShaded() bool {
	return (sf & StyleFlagShaded) != 0
}

func (sf *StyleFlags) SetIsShaded(id bool) {
	if id {
		*sf |= StyleFlagShaded
	} else {
		*sf &= ^StyleFlagShaded
	}
}

const StyleFlagNone StyleFlags = 0

const (
	StyleFlagBold StyleFlags = 1 << iota
	StyleFlagDim
	StyleFlagShaded
	StyleFlagBlink
	StyleFlagUnderline
	StyleFlagItalic
	StyleFlagStrikeThrough
)

// Attrs have side effects in the Logos system;
// for example, the lone cursor element (one with AttrFlagIsCursor set)
// is where most key events are sent to.
type Attrs struct {
	Parent Elem
	AttrFlags
	Other []KVPair
}

func (tt *Attrs) GetAttrs() *Attrs {
	return tt
}

func (tt *Attrs) GetParent() Elem {
	return tt.Parent
}

func (tt *Attrs) SetParent(p Elem) {
	if tt.Parent != nil && tt.Parent != p {
		panic("parent already set")
	}
	tt.Parent = p
}

func (tt *Attrs) GetIsCursor() bool {
	return (tt.AttrFlags & AttrFlagIsCursor) != 0
}

func (tt *Attrs) SetIsCursor(ic bool) {
	if ic {
		tt.AttrFlags |= AttrFlagIsCursor
	} else {
		tt.AttrFlags &= ^AttrFlagIsCursor
	}
	tt.SetIsDirty(true)
}

func (tt *Attrs) GetIsDirty() bool {
	return (tt.AttrFlags & AttrFlagIsDirty) != 0
}

func (tt *Attrs) SetIsDirty(id bool) {
	if id {
		tt.AttrFlags |= AttrFlagIsDirty
		if tt.Parent != nil {
			tt.Parent.SetIsDirty(true)
		}
	} else {
		tt.AttrFlags &= ^AttrFlagIsDirty
	}
}

func (tt *Attrs) GetIsOccluded() bool {
	return (tt.AttrFlags & AttrFlagIsOccluded) != 0
}

func (tt *Attrs) SetIsOccluded(ic bool) {
	if ic {
		tt.AttrFlags |= AttrFlagIsOccluded
	} else {
		tt.AttrFlags &= ^AttrFlagIsOccluded
	}
	tt.SetIsDirty(true)
}

func (tt *Attrs) Merge(ot *Attrs) {
	if ot.Parent != nil {
		tt.Parent = ot.Parent
	}
	tt.AttrFlags |= ot.AttrFlags
	tt.Other = ot.Other // TODO merge by key.
}

// ----------------------------------------
// AttrFlags

// NOTE: AttrFlags are merged with a simple or-assign op.
type AttrFlags uint32

func (af AttrFlags) GetAttrFlags() AttrFlags {
	return af
}

const AttrFlagNone AttrFlags = 0

const (
	AttrFlagIsCursor   AttrFlags = 1 << iota // is current cursor
	AttrFlagIsSelected                       // is selected (among possibly others)
	AttrFlagIsOccluded                       // is hidden due to stack
	AttrFlagIsDirty                          // is dirty (not yet used)
)

type KVPair struct {
	Key   string
	Value interface{}
}

// ----------------------------------------
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
			+----------+
			x   0
		*/
		minX = 0
	} else {
		/*
			     View
			     +----------+
			[____|__Elem]   |
			     +----------+
			0    x
		*/
		minX = elo.X
	}
	if els.Width <= vws.Width+elo.X {
		/*
			     View
			     +----------+
			[____|__Elem]   |
			     +----------+
				        W   w+x
		*/
		maxX = els.Width
	} else {
		/*
			View
			+----------+
			|   [Elem__|____]
			+----------+
			           w+x  W
		*/
		maxX = vws.Width + elo.X
	}
	if elo.Y < 0 {
		minY = 0
	} else {
		minY = elo.Y
	}
	if els.Height <= vws.Height+elo.Y {
		maxY = els.Height
	} else {
		maxY = vws.Height + elo.Y
	}
	return
}

// ----------------------------------------
// Misc simple types

type Padding struct {
	Left   int
	Top    int
	Right  int
	Bottom int
}

func (pd Padding) GetPadding() Padding {
	return pd
}

// A border can only have width 0 or 1, and is part of the padding.
// Each string should represent a character of width 1.
type Border struct {
	Corners   [4]string // starts upper-left and clockwise, "" draws no corner.
	TopLine   []string  // nil if no top border.
	BotLine   []string  // nil if no bottom border.
	LeftLine  []string  // nil if no left border.
	RightLine []string  // nil if no right border.
}

func DefaultBorder() Border {
	return Border{
		Corners: [4]string{
			string(tcell.RuneULCorner),
			string(tcell.RuneURCorner),
			string(tcell.RuneLRCorner),
			string(tcell.RuneLLCorner),
		},
		TopLine:   []string{string(tcell.RuneHLine)},
		BotLine:   []string{string(tcell.RuneHLine)},
		LeftLine:  []string{string(tcell.RuneVLine)},
		RightLine: []string{string(tcell.RuneVLine)},
	}
}

func LeftBorder() Border {
	return Border{
		Corners: [4]string{
			string("\u2553"),
			"",
			"",
			string("\u2559"),
		},
		LeftLine: []string{
			string("\u2551"),
		},
	}
}

func orSpace(chr string) string {
	if chr == "" {
		return " "
	} else {
		return chr
	}
}

func (br *Border) GetCorner(i int) string {
	if br == nil {
		return " "
	} else {
		return orSpace(br.Corners[i])
	}
}

func (br *Border) TLCorner() string {
	return br.GetCorner(0)
}

func (br *Border) TRCorner() string {
	return br.GetCorner(1)
}

func (br *Border) BRCorner() string {
	return br.GetCorner(2)
}

func (br *Border) BLCorner() string {
	return br.GetCorner(3)
}

func (br *Border) TopBorder(x int) string {
	if br == nil || br.TopLine == nil {
		return " "
	} else {
		return br.TopLine[x%len(br.TopLine)]
	}
}

func (br *Border) BottomBorder(x int) string {
	if br == nil || br.BotLine == nil {
		return " "
	} else {
		return br.BotLine[x%len(br.BotLine)]
	}
}

func (br *Border) LeftBorder(y int) string {
	if br == nil || br.LeftLine == nil {
		return " "
	} else {
		return br.LeftLine[y%len(br.LeftLine)]
	}
}

func (br *Border) RightBorder(y int) string {
	if br == nil || br.RightLine == nil {
		return " "
	} else {
		return br.RightLine[y%len(br.RightLine)]
	}
}

type Size struct {
	Width  int
	Height int // -1 if not set.
}

func (sz Size) String() string {
	return fmt.Sprintf("{%d,%d}", sz.Width, sz.Height)
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

func (crd *Coord) SetCoord(nc Coord) {
	*crd = nc
}

func (crd Coord) IsNonNegative() bool {
	return 0 <= crd.X && 0 <= crd.Y
}

func (crd Coord) Neg() Coord {
	return Coord{
		X: -crd.X,
		Y: -crd.Y,
	}
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
