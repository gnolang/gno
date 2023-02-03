package logos

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ----------------------------------------
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

func (st *Stack) StringIndented(indent string) string {
	elines := []string{}
	eindent := indent + "    "
	for _, elem := range st.Elems {
		elines = append(elines, eindent+elem.StringIndented(eindent))
	}

	return fmt.Sprintf("Stack%v@%p\n%s",
		st.Size,
		st,
		strings.Join(elines, "\n"))
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

// Draw the rendered layers onto the view.  Any dimming of
// occluded layers must actually stretch in all directions
// infinitely (since we can scroll beyond the bounds of any
// view and we expect the dimming effect to carry while we
// scroll), so the entire view is dimmed first, and then the
// upper-most layer is drawn.
func (st *Stack) Draw(offset Coord, view View) {
	// Draw bottom layers.
	if 1 < len(st.Elems) {
		for _, elem := range st.Elems[:len(st.Elems)-1] {
			loffset := offset.Sub(elem.GetCoord())
			elem.Draw(loffset, view)
		}
	}
	if 0 < len(st.Elems) {
		last := st.Elems[len(st.Elems)-1]
		loffset := offset.Sub(last.GetCoord())
		// Draw occlusion screen on view.
		for y := 0; y < view.Bounds.Height; y++ {
			for x := 0; x < view.Bounds.Width; x++ {
				vcell := view.GetCell(x, y)
				inBounds := IsInBounds(x, y,
					loffset.Neg(),
					last.GetSize())
				if inBounds {
					// Reset unsets residual "occluded",
					// "cursor", and other attributes from the
					// previous layer which are no longer
					// relevant.
					vcell.Reset()
				} else {
					vcell.SetIsOccluded(true)
				}
			}
		}
		// Draw last (top) layer.
		last.Draw(loffset, view)
	} else {
		// Draw occlusion screen on view.
		for y := 0; y < view.Bounds.Height; y++ {
			for x := 0; x < view.Bounds.Width; x++ {
				vcell := view.GetCell(x, y)
				vcell.SetIsOccluded(true)
			}
		}
	}
}

func (st *Stack) ProcessEventKey(ev *EventKey) bool {
	// An empty *Stack is inert.
	if len(st.Page.Elems) == 0 {
		return false
	}
	// Try to let the last layer handle it.
	last := st.Page.Elems[len(st.Page.Elems)-1]
	if last.ProcessEventKey(ev) {
		return true
	}
	// Maybe it's something for the stack.
	switch ev.Key() {
	case tcell.KeyEsc:
		if 1 < len(st.Page.Elems) {
			// Pop the last layer.
			st.Elems = st.Elems[:len(st.Elems)-1]
			st.Cursor--
			st.SetIsDirty(true)

			return true
		} else {
			// Let the last layer stick around.
			return false
		}
	default:
		return false
	}
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
