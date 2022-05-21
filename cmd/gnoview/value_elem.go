package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/gnolang/gno"
	"github.com/gnolang/gno/logos"
)

var tstyle = &logos.Style{ // text
	// Background: tcell.ColorWhite,
	CursorStyle: &logos.Style{
		Background: tcell.ColorYellow,
	},
}

var istyle = &logos.Style{ // elem in list
	// Background: tcell.ColorWhite,
	Padding: logos.Padding{2, 0, 0, 0},
	Border:  logos.LeftBorder(),
	CursorStyle: &logos.Style{
		Padding:    logos.Padding{2, 0, 0, 0},
		Border:     logos.LeftBorder(),
		Background: tcell.ColorYellow,
	},
}

var bstyle = &logos.Style{ // box style
	// Background: tcell.ColorWhite,
	Padding: logos.Padding{2, 0, 2, 1},
	Border:  logos.DefaultBorder(),
	CursorStyle: &logos.Style{
		Padding:    logos.Padding{2, 0, 2, 1},
		Border:     logos.LeftBorder(),
		Background: tcell.ColorYellow,
	},
}

func MakeTypedValueElem(tv *gno.TypedValue) logos.Elem {
	// make a buffered page.
	page := logos.NewPage("", 84, true, istyle)
	if tv.IsUndefined() {
		// add "undefined"
		elemv := logos.NewTextElem("undefined", tstyle)
		page.AppendElem(elemv)
	} else {
		// add type info unless primitive.
		if _, prim := gno.BaseOf(tv.T).(gno.PrimitiveType); !prim {
			elemt := logos.NewTextElem(
				fmt.Sprintf("%s: %s",
					tv.T.Kind().String(),
					tv.T.String()),
				tstyle)
			page.AppendElem(elemt)
		}
		// add elements depending on type.
		switch gno.BaseOf(tv.T).(type) {
		case gno.PrimitiveType:
			elemv := logos.NewTextElem(tv.String(), tstyle)
			page.AppendElem(elemv)
		case *gno.PackageType:
			if tv.V == nil {
				elemv := logos.NewTextElem("nil", tstyle)
				page.AppendElem(elemv)
			} else {
				elemv1 := logos.NewTextElem("Values:", tstyle)
				page.AppendElem(elemv1)
				elemv2 := MakeValueElem(tv.V, nil)
				page.AppendElem(elemv2)
			}
		case *gno.StructType:
			if tv.V == nil {
				elemv := logos.NewTextElem("zero", tstyle)
				page.AppendElem(elemv)
			} else {
				elemv2 := MakeValueElem(tv.V, nil)
				page.AppendElem(elemv2)
			}
		case *gno.TypeType:
			elemv1 := logos.NewTextElem("String: "+tv.GetType().String(), tstyle)
			page.AppendElem(elemv1)
			elemv2 := logos.NewTextElem("TypeID: #"+tv.GetType().TypeID().String(), tstyle)
			page.AppendElem(elemv2)
		default:
			// XXX fix.
			elemv := logos.NewTextElem("XXX: "+tv.String(), nil)
			page.AppendElem(elemv)
		}
	}
	// measure page and return buffered view.
	tve := logos.NewBufferedElemView(page, logos.Size{})
	return tve
}

func MakeValueElem(value gno.Value, style *logos.Style) logos.Elem {
	// make a buffered page.
	page := logos.NewPage("", 84, true, style)
	// add elements depending on type.
	switch cv := value.(type) {
	case nil:
		// add "undefined".
		elemt := logos.NewTextElem("undefined", tstyle)
		page.AppendElem(elemt)
	case *gno.PackageValue:
		// add heading.
		elemt := logos.NewTextElem("Package values:", tstyle)
		page.AppendElem(elemt)
		// add package values.
		for i := 0; i < len(cv.Values); i++ {
			ev := &cv.Values[i]
			eleme := MakeTypedValueElem(ev)
			page.AppendElem(eleme)
		}
	case *gno.StructValue:
		oinfo := cv.GetObjectInfo()
		// add object id.
		elemo := logos.NewTextElem("ObjectID: "+
			oinfo.GetObjectID().String(), tstyle)
		page.AppendElem(elemo)
		// add owner id.
		elemow := logos.NewTextElem("Owner: "+
			oinfo.GetOwnerID().String(), tstyle)
		page.AppendElem(elemow)
		// add mod time.
		elemfz := logos.NewTextElem(
			fmt.Sprintf(
				"ModTime: %d, RefCount: %d",
				oinfo.GetModTime(),
				oinfo.GetRefCount()),
			tstyle)
		page.AppendElem(elemfz)
		// add fields heading.
		elemt := logos.NewTextElem("Fields:", tstyle)
		page.AppendElem(elemt)
		// add struct fields.
		for i := 0; i < len(cv.Fields); i++ {
			ev := &cv.Fields[i]
			eleme := MakeTypedValueElem(ev)
			page.AppendElem(eleme)
		}
	default:
		panic("should not happen")
	}
	// measure page and return buffered view.
	bpv := logos.NewBufferedElemView(page, logos.Size{})
	return bpv
}
