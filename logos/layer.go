package logos

// A terminal character cell.
type Cell struct {
	Character  string // 1 unicode character, or " ".
	Foreground Color
	Background Color
	Flags      Flags
	Elem       Elem // reference to element
}

func (cc *Cell) SetCell(oc *Cell) {
	*cc = *oc
}

func (cc *Cell) SetValue(chs string, w int, st Style, el Elem) {
	cc.Character = chs
	cc.Foreground = st.Foreground
	cc.Background = st.Background
	cc.Flags = st.Flags
	cc.Elem = el
}

// A view onto a page.
// Somewhat like a slice onto an array
// (as a view is onto a page),
// except cells are allocated here.
type BufferedPageView struct {
	Coord
	Size
	Style
	*Page
	Offset  Coord // within page for pagination
	*Buffer       // view's internal draw screen
}

// Renders the page onto the internal buffer.
// Assumes buffered page view's page was already rendered.
// TODO: this function could be optimized to reduce
// redundant background cell modifications.
func (bpv *BufferedPageView) Render() {
	// First, draw page background style.
	style := bpv.Style
	buffer := bpv.Buffer
	for x := 0; x < buffer.Size.Width; x++ {
		for y := 0; y < buffer.Size.Height; y++ {
			cell := buffer.GetCell(x, y)
			cell.Foreground = style.Foreground
			cell.Background = style.Foreground
			cell.Flags = style.Flags
		}
	}
	// Then, render and draw page.
	bpv.Page.Render()
	bpv.Page.Draw(bpv.Offset, buffer.NewView(Coord{}))
}

func (bpv *BufferedPageView) Draw(offset Coord, view View) {
	minX, maxX, minY, maxY := computeIntersection(bpv.Size, offset, view.Size)
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			bcell := bpv.Buffer.GetCell(x, y)
			vcell := view.GetCell(x-minX, y-minY)
			vcell.SetCell(bcell)
		}
	}
}

type Screen struct {
	BufferedPageView
}
