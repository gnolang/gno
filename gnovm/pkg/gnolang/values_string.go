package gnolang

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

const (
	// defaultSeenValuesSize indicates the maximum anticipated depth of the stack when printing a Value type.
	defaultSeenValuesSize = 32

	// nestedLimit indicates the maximum nested level when printing a deeply recursive value.
	// if this increases significantly a map should be used instead
	nestedLimit = 10
)

type seenValues struct {
	values []Value
}

func (sv *seenValues) Put(v Value) bool {
	if len(sv.values) >= nestedLimit {
		return false
	}

	sv.values = append(sv.values, v)
	return true
}

func (sv *seenValues) IndexOf(v Value) int {
	for i, vv := range sv.values {
		if vv == v {
			return i
		}
	}

	return -1
}

// Pop should be called by using a defer after each Put.
// Consider why this is necessary:
//   - we are printing an array of structs
//   - each invocation of struct.ProtectedString adds the value to the seenValues
//   - without calling Pop before exiting struct.ProtectedString, the next call to
//     struct.ProtectedString in the array.ProtectedString loop will not result in the value
//     being printed if the value has already been print
//   - this is NOT recursion and SHOULD be printed
func (sv *seenValues) Pop() {
	sv.values = sv.values[:len(sv.values)-1]
}

func newSeenValues() *seenValues {
	return &seenValues{
		values: make([]Value, 0, defaultSeenValuesSize),
	}
}

func (sv StringValue) String() string {
	return strconv.Quote(string(sv))
}

func (biv BigintValue) String() string {
	return biv.V.String()
}

func (bdv BigdecValue) String() string {
	s := bdv.V.FloatString(10)
	// Trim trailing zeros after the decimal point, but keep at least one
	// decimal digit so bigdec values are visually distinct from integers.
	if strings.ContainsRune(s, '.') {
		s = strings.TrimRight(s, "0")
		// Keep at least one digit after the decimal point.
		if s[len(s)-1] == '.' {
			s += "0"
		}
	}
	return s
}

func (dbv DataByteValue) String() string {
	return fmt.Sprintf("(%0X)", (dbv.GetByte()))
}

// protectedStringOf renders v through the writer-based formatter and
// returns the accumulated bytes. The meteredWriter buffers internally and
// flushes a single chunk into the bytes.Buffer, so no per-shape size
// hint is needed (the debug paths pass no machine, so no gas is charged).
func protectedStringOf(v protectedWriter, seen *seenValues) string {
	var b bytes.Buffer
	mw := newUnmeteredWriter(&b)
	defer mw.Release()
	v.WriteProtected(mw, seen)
	// Flush explicitly (not deferred): b is read below via b.String(), and a
	// deferred Flush would run only after the return value is evaluated,
	// dropping the last buffered chunk. Streaming callers that hand the buffer
	// to a downstream consumer (uversePrint, Fprint) defer Flush instead.
	mw.Flush()
	return b.String()
}

func (av *ArrayValue) String() string {
	return av.ProtectedString(newSeenValues())
}

func (av *ArrayValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(av, seen)
}

func (sv *SliceValue) String() string {
	return sv.ProtectedString(newSeenValues())
}

func (sv *SliceValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(sv, seen)
}

func (pv PointerValue) String() string {
	return pv.ProtectedString(newSeenValues())
}

func (pv PointerValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(pv, seen)
}

func (sv *StructValue) String() string {
	return sv.ProtectedString(newSeenValues())
}

func (sv *StructValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(sv, seen)
}

func (fv *FuncValue) String() string {
	name := string(fv.Name)
	if fv.Type == nil {
		return fmt.Sprintf("incomplete-func ?%s(?)?", name)
	}
	if name == "" {
		return fmt.Sprintf("%s{...}", fv.Type.String())
	}
	return name
}

func (bmv *BoundMethodValue) String() string {
	name := bmv.Func.Name
	var (
		recvT   = "?"
		params  = "?"
		results = "(?)"
	)
	if ft, ok := bmv.Func.Type.(*FuncType); ok {
		recvT = ft.Params[0].Type.String()
		params = FieldTypeList(ft.Params).StringForFunc()
		if len(results) > 0 {
			results = FieldTypeList(ft.Results).StringForFunc()
			results = "(" + results + ")"
		}
	}
	return fmt.Sprintf("<%s>.%s(%s)%s",
		recvT, name, params, results)
}

func (mv *MapValue) String() string {
	return mv.ProtectedString(newSeenValues())
}

func (mv *MapValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(mv, seen)
}

func (tv TypeValue) String() string {
	return fmt.Sprintf("typeval{%s}",
		tv.Type.String())
}

func (pv *PackageValue) String() string {
	return fmt.Sprintf("package(%s %s)", pv.PkgName, pv.PkgPath)
}

func (b *Block) String() string {
	return b.StringIndented("    ")
}

func (b *Block) StringIndented(indent string) string {
	source := toString(b.Source)
	if len(source) > 32 {
		source = source[:32] + "..."
	}
	lines := make([]string, 0, 3)
	lines = append(lines,
		fmt.Sprintf("Block(ID:%v,Addr:%p,Source:%s,Parent:%p)",
			b.ObjectInfo.ID, b, source, b.Parent)) // XXX Parent may be RefValue{}.
	if b.Source != nil {
		if _, ok := b.Source.(RefNode); ok {
			lines = append(lines,
				fmt.Sprintf("%s(RefNode names not shown)", indent))
		} else {
			types := b.Source.GetStaticBlock().Types
			for i, n := range b.Source.GetBlockNames() {
				if len(b.Values) <= i {
					lines = append(lines,
						fmt.Sprintf("%s%s: undefined static:%s", indent, n, types[i]))
				} else {
					lines = append(lines,
						fmt.Sprintf("%s%s: %s static:%s",
							indent, n, b.Values[i].String(), types[i]))
				}
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (rv RefValue) String() string {
	if rv.PkgPath == "" {
		return fmt.Sprintf("ref(%v)",
			rv.ObjectID)
	}
	return fmt.Sprintf("ref(%s)",
		rv.PkgPath)
}

func (hiv *HeapItemValue) String() string {
	return fmt.Sprintf("heapitem(%v)",
		hiv.Value)
}

// ----------------------------------------
// *TypedValue.Sprint

// ImplError returns true if the TypedValue's type implements the error interface.
func (tv *TypedValue) ImplError() bool {
	return IsImplementedBy(gErrorType, tv.T)
}

// for print() and println().
func (tv *TypedValue) Sprint(m *Machine) string {
	var b bytes.Buffer
	tv.Fprint(&b, m)
	return b.String()
}

func (tv *TypedValue) ProtectedSprint(seen *seenValues, considerDeclaredType bool) string {
	var b bytes.Buffer
	mw := newUnmeteredWriter(&b)
	defer mw.Release()
	writeProtectedSprint(mw, *tv, seen, considerDeclaredType)
	mw.Flush() // explicit, not deferred — b is read below; see protectedStringOf.
	return b.String()
}

// ----------------------------------------
// TypedValue.String()

// For gno debugging/testing.
func (tv TypedValue) String() string {
	return tv.ProtectedString(newSeenValues())
}

func (tv TypedValue) ProtectedString(seen *seenValues) string {
	return protectedStringOf(tv, seen)
}
