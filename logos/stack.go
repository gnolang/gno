package logos

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

//----------------------------------------
// Stack

// A Stack is like a Page, but it only highlights the top
// element, and dims, occludes, or hides lower elements.  A
// Stack is therefore ideal for showing modal views.  NOTE:
// While most applications shouldn't, it's perfectly fine to
// embed Stacks within Stacks, or layer them on top within a
// stack.
type Stack struct {
	Page // a Stack has the same fields as a page.
}

func NewStack(size Size) *Stack {
	return &Stack{
		Page: Page{
			Size:   size, // dontcare.
			Elems:  nil,  // nil layers.
			Cursor: -1,
		},
	}
}

func (st *Stack) String() string {
	return fmt.Sprintf("Stack%v{%d}@%p",
		st.Size,
		len(st.Elems),
		st)
}

func (st *Stack) PushLayer(layer Elem) {
	layer.SetParent(st)
	st.Elems = append(st.Elems, layer)
	st.Cursor++
	st.SetIsDirty(true)
}

// A Stack's size is simply determined by its .Size.
func (st *Stack) Measure() Size {
	return st.Size
}

// A Stack's render function behaves the same as a Page's;
// it renders its elements (here, its layers).
func (st *Stack) Render() (updated bool) {
	return st.Page.Render()
}

// Draw the rendered layers onto the view.
func (st *Stack) Draw(offset Coord, view View) {
	style := st.Style
	minX, maxX, minY, maxY :=
		computeIntersection(st.Size, offset, view.Bounds)
	// First, draw page background style.
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			xo, yo := x-offset.X, y-offset.Y
			vcell := view.GetCell(xo, yo)
			if style.Border.HasBorder {
				// draw area and border
				if x == 0 {
					if y == 0 {
						vcell.SetValue(string(tcell.RuneULCorner), 1, style, nil)
					} else if y == st.Size.Height-1 {
						vcell.SetValue(string(tcell.RuneLLCorner), 1, style, nil)
					} else {
						vcell.SetValue(string(tcell.RuneVLine), 1, style, nil)
					}
				} else if x == st.Size.Width-1 {
					if y == 0 {
						vcell.SetValue(string(tcell.RuneURCorner), 1, style, nil)
					} else if y == st.Size.Height-1 {
						vcell.SetValue(string(tcell.RuneLRCorner), 1, style, nil)
					} else {
						vcell.SetValue(string(tcell.RuneVLine), 1, style, nil)
					}
				} else if y == 0 {
					vcell.SetValue(string(tcell.RuneHLine), 1, style, nil)
				} else if y == st.Size.Height-1 {
					vcell.SetValue(string(tcell.RuneHLine), 1, style, nil)
				} else {
					vcell.SetValue(" ", 1, style, nil)
				}
			} else {
				// draw area but no border.
				vcell.SetValue(" ", 1, style, nil)
			}
		}
	}
	// Then, draw layers.
	if len(st.Elems) > 0 {
		if true {
			// Draw bottom layers.
			if len(st.Elems) > 1 {
				for _, layer := range st.Elems[:len(st.Elems)-1] {
					loffset := offset.Sub(layer.GetCoord())
					layer.Draw(loffset, view)
				}
			}
			// Draw occlusion layer.
			for y := minY; y < maxY; y++ {
				for x := minX; x < maxX; x++ {
					xo, yo := x-offset.X, y-offset.Y
					vcell := view.GetCell(xo, yo)
					vcell.SetIsShaded(true)
				}
			}
		}
		// Draw top layer.
		llayer := st.Elems[len(st.Elems)-1]
		loffset := offset.Sub(llayer.GetCoord())
		llayer.Draw(loffset, view)
	}
	if debug {
		debug.Println("sleeping after drawing page")
		time.Sleep(time.Second)
	}
}

func (st *Stack) ProcessEventKey(ev *EventKey) bool {
	if len(st.Page.Elems) == 0 {
		return false
	}
	// XXX layer operations.
	last := st.Page.Elems[len(st.Page.Elems)-1]
	return last.ProcessEventKey(ev)
}

// Traverses the inclusive ancestors of elem and returns the
// first *Stack encountered.  The purpose of this function is
// to find where to push new layers and modal elements for
// drawing.
func StackOf(elem Elem) *Stack {
	for elem != nil {
		fmt.Println("StackOf", elem)
		if st, ok := elem.(*Stack); ok {
			return st
		} else {
			elem = elem.GetParent()
		}
	}
	return nil // no stack
}
