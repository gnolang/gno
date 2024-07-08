package gengo

import (
	"fmt"
	"reflect"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/libs/press"
)

// shortcut
var _fmt = fmt.Sprintf

func PrintIntEncoder(p *press.Press, ref string) {
	p.Pl("{").I(func(p *press.Press) {
		p.Pl("var buf[10]byte")
		p.Pl("n := binary.PutVarint(buf[:], %v)", ref)
		p.Pl("_, err = w.Write(buf[0:n])")
	}).Pl("}")
}

func PrintStructFieldEncoder(p *press.Press, ref string, info amino.FieldInfo) {
	name := info.Name
	fref := ref + "." + name // TODO document restriction on name types to make naive "+" possible.
	done := p.RandID("done")
	cond := printStructFieldSkipCond(p, fref, info)
	p.Pl("{").I(func(p *press.Press) {
		p.Pl("// Struct field %v", name)
		p.Pl("// Maybe skip?")
		p.Pl("// if (%v) {", cond).I(func(p *press.Press) {
			p.Pl("goto %v", done)
		}).Pl("}")
		p.Pl("pos1 := w.Len()")
		p.Pl("// Write field number & typ3")
		p.Pl("// TODO")
		p.Pl("pos2 := w.Len()")
		p.Pl("// Write field value")
		// XXX PrintValueEncoder(p, fref, info.Type)
		p.Pl("XXX PrintValueEncoder()")
		p.Pl("pos3 := w.Len()")
		// Maybe skip the writing of zero structs unless also WriteEmpty.
		if info.Type.Kind() == reflect.Ptr && !info.WriteEmpty {
			p.Pl("if (pos2 == pos3-1 && w.PeekLastByte() == 0x00) {").I(func(p *press.Press) {
				p.Pl("w.Truncate(pos1)")
			}).Pl("}")
		}
	}).Pl("}")
}

func printStructFieldSkipCond(p *press.Press, fref string, info amino.FieldInfo) string {
	// If the value is nil or empty, do not encode.
	// Amino "zero" structs are not yet well defined/understood.
	// Field values that are zero structs (and !WriteEmpty) are not skipped
	// here, but later in PrintStructFieldEncoder() upon inspecting the number
	// of written bytes.
	switch info.Type.Kind() {
	case reflect.Ptr:
		return _fmt("%v == nil", fref)
	case reflect.Bool:
		return _fmt("%v == false", fref)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return _fmt("%v == 0", fref)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return _fmt("%v == 0", fref)
	case reflect.String:
		return _fmt("len(%v) == 0", fref)
	case reflect.Chan, reflect.Map, reflect.Slice:
		return _fmt("%v == nil || len(%v) == 0", fref, fref)
	case reflect.Func, reflect.Interface:
		return _fmt("%v == nil", fref)
	default:
		return "true"
	}
}
