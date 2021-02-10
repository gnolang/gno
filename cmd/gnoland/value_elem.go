package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/gnolang/gno"
	"github.com/gnolang/gno/logos"
)

var zstyle = logos.Style{} // zero style.

var istyle = logos.Style{ // elem in list
	Padding: logos.Padding{2, 0, 0, 0},
	// Background: tcell.ColorYellow,
	Border: logos.LeftBorder(),
}

var bstyle = logos.Style{ // box style
	Padding:    logos.Padding{2, 0, 2, 1},
	Background: tcell.ColorWhite,
	Border:     logos.DefaultBorder(),
}

func MakeTypedValueElem(tv *gno.TypedValue) logos.Elem {
	// make a buffered page.
	page := logos.NewPage("", 84, true, istyle)
	if tv.IsUndefined() {
		// add "undefined"
		elemv := logos.NewTextElem("undefined", zstyle)
		page.AppendElem(elemv)
	} else {
		// add type info unless primitive.
		if _, prim := gno.BaseOf(tv.T).(gno.PrimitiveType); !prim {
			elemt := logos.NewTextElem(
				fmt.Sprintf("%s: %s",
					tv.T.Kind().String(),
					tv.T.String()),
				zstyle)
			page.AppendElem(elemt)
		}
		// add elements depending on type.
		switch gno.BaseOf(tv.T).(type) {
		case gno.PrimitiveType:
			elemv := logos.NewTextElem(tv.String(), zstyle)
			page.AppendElem(elemv)
		case *gno.PackageType:
			if tv.V == nil {
				elemv := logos.NewTextElem("nil", zstyle)
				page.AppendElem(elemv)
			} else {
				elemv1 := logos.NewTextElem("Values:", zstyle)
				page.AppendElem(elemv1)
				elemv2 := MakeValueElem(tv.V, zstyle)
				page.AppendElem(elemv2)
			}
		case *gno.StructType:
			if tv.V == nil {
				elemv := logos.NewTextElem("zero", zstyle)
				page.AppendElem(elemv)
			} else {
				elemv2 := MakeValueElem(tv.V, zstyle)
				page.AppendElem(elemv2)
			}
		case *gno.TypeType:
			elemv1 := logos.NewTextElem("String: "+tv.GetType().String(), zstyle)
			page.AppendElem(elemv1)
			elemv2 := logos.NewTextElem("TypeID: #"+tv.GetType().TypeID().String(), zstyle)
			page.AppendElem(elemv2)
		default:
			// XXX fix.
			elemv := logos.NewTextElem("XXX: "+tv.String(), zstyle)
			page.AppendElem(elemv)
		}
	}
	// measure page and return buffered view.
	tve := logos.NewBufferedElemView(page, logos.Size{})
	return tve
}

func MakeValueElem(value gno.Value, style logos.Style) logos.Elem {
	// make a buffered page.
	page := logos.NewPage("", 84, true, style)
	// add elements depending on type.
	switch cv := value.(type) {
	case nil:
		// add "undefined".
		elemt := logos.NewTextElem("undefined", zstyle)
		page.AppendElem(elemt)
	case *gno.PackageValue:
		// add heading.
		elemt := logos.NewTextElem("Package values:", zstyle)
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
			oinfo.GetObjectID().String(), zstyle)
		page.AppendElem(elemo)
		// add owner id.
		elemow := logos.NewTextElem("Owner: "+
			oinfo.GetOwnerID().String(), zstyle)
		page.AppendElem(elemow)
		// add mod time.
		elemfz := logos.NewTextElem(
			fmt.Sprintf(
				"ModTime: %d, RefCount: %d",
				oinfo.GetModTime(),
				oinfo.GetRefCount()),
			zstyle)
		page.AppendElem(elemfz)
		// add fields heading.
		elemt := logos.NewTextElem("Fields:", zstyle)
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
