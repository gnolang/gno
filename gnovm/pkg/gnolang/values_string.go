package gnolang

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

const (
	// defaultSeenValuesSize indicates the maximum anticipated depth of the stack when printing a Value type.
	defaultSeenValuesSize = 32

	// nestedLimit indicates the maximum nested level when printing a deeply recursive value.
	// if this increases significantly a map should be used instead
	nestedLimit = 10

	// printLimit is the maximum number of elements (or bytes for byte-backed data)
	// to display in string representations of arrays, slices, and maps.
	printLimit = 256

	// printOutputLimit is the maximum length of any string produced by the
	// String()/Sprint() entry points. Every write goes through boundedBuilder,
	// which enforces this cap and lets renderers stop descending once it is
	// reached. This bounds the total work of printing a value to
	// O(printOutputLimit) regardless of how deeply or widely it is nested,
	// preventing combinatorial blow-ups (and the native allocations they would
	// otherwise cause) from print/println.
	printOutputLimit = 64_000

	// truncatedSuffix is appended once printOutputLimit is reached.
	truncatedSuffix = "...(truncated)"
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
//   - each invocation of struct.writeBare adds the value to the seenValues
//   - without calling Pop before exiting struct.writeBare, the next call to
//     struct.writeBare in the array.writeBare loop will not result in the value
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

// boundedBuilder accumulates a value's string representation while enforcing a
// hard cap of printOutputLimit bytes. Once the cap is reached it appends a
// single truncation marker and ignores all further writes; renderers consult
// done() to stop iterating, which is what keeps printing bounded.
type boundedBuilder struct {
	b         strings.Builder
	truncated bool
}

func newBoundedBuilder() *boundedBuilder {
	return &boundedBuilder{}
}

// writeString appends s, or as much of it as fits, marking the output
// truncated once the limit is reached.
func (w *boundedBuilder) writeString(s string) {
	if w.truncated {
		return
	}
	if w.b.Len()+len(s) > printOutputLimit {
		if avail := printOutputLimit - w.b.Len(); avail > 0 {
			w.b.WriteString(s[:avail])
		}
		w.b.WriteString(truncatedSuffix)
		w.truncated = true
		return
	}
	w.b.WriteString(s)
}

func (w *boundedBuilder) writeByte(c byte) {
	if w.truncated {
		return
	}
	if w.b.Len()+1 > printOutputLimit {
		w.b.WriteString(truncatedSuffix)
		w.truncated = true
		return
	}
	w.b.WriteByte(c)
}

// done reports whether the output limit has been reached. Renderers use it to
// stop iterating once no further output can be produced.
func (w *boundedBuilder) done() bool {
	return w.truncated
}

func (w *boundedBuilder) String() string {
	return w.b.String()
}

// writeSep writes the element separator before all but the first element and
// reports whether rendering should stop because the output cap was reached.
func (w *boundedBuilder) writeSep(i int) (stop bool) {
	if w.done() {
		return true
	}
	if i > 0 {
		w.writeByte(',')
	}
	return false
}

func (sv StringValue) String() string {
	return strconv.Quote(string(sv))
}

func (biv BigintValue) String() string {
	return biv.V.String()
}

func (bdv BigdecValue) String() string {
	return bdv.V.String()
}

func (dbv DataByteValue) String() string {
	return fmt.Sprintf("(%0X)", (dbv.GetByte()))
}

func (av *ArrayValue) String() string {
	w := newBoundedBuilder()
	av.writeBare(w, newSeenValues())
	return w.String()
}

// writeBare renders the array's "array[...]" form into w.
func (av *ArrayValue) writeBare(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if i := seen.IndexOf(av); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}
	if !seen.Put(av) {
		w.writeString("...")
		return
	}
	defer seen.Pop()

	if av.Data == nil {
		if len(av.List) > printLimit {
			w.writeString(fmt.Sprintf("array[...(%d elements)]", len(av.List)))
			return
		}
		w.writeString("array[")
		for i := range av.List {
			if w.writeSep(i) {
				break
			}
			av.List[i].writeWrapped(w, seen)
		}
		// NOTE: we may want to unify the representation,
		// but for now tests expect this to be different.
		// This may be helpful for testing implementation behavior.
		w.writeByte(']')
		return
	}
	if len(av.Data) > printLimit {
		w.writeString(fmt.Sprintf("array[0x%X...(%d)]", av.Data[:printLimit], len(av.Data)))
		return
	}
	w.writeString(fmt.Sprintf("array[0x%X]", av.Data))
}

func (sv *SliceValue) String() string {
	w := newBoundedBuilder()
	sv.writeBare(w, newSeenValues())
	return w.String()
}

// writeBare renders the slice's "slice[...]" form into w.
func (sv *SliceValue) writeBare(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if sv.Base == nil {
		w.writeString("nil-slice")
		return
	}
	if i := seen.IndexOf(sv); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}
	if ref, ok := sv.Base.(RefValue); ok {
		w.writeString(fmt.Sprintf("slice[%v]", ref))
		return
	}
	if !seen.Put(sv) {
		w.writeString("...")
		return
	}
	defer seen.Pop()

	vbase := sv.Base.(*ArrayValue)
	if vbase.Data == nil {
		if sv.Length > printLimit {
			w.writeString(fmt.Sprintf("slice[...(%d elements)]", sv.Length))
			return
		}
		w.writeString("slice[")
		for i := 0; i < sv.Length; i++ {
			if w.writeSep(i) {
				break
			}
			vbase.List[sv.Offset+i].writeWrapped(w, seen)
		}
		w.writeByte(']')
		return
	}
	if sv.Length > printLimit {
		w.writeString(fmt.Sprintf("slice[0x%X...(%d)]", vbase.Data[sv.Offset:sv.Offset+printLimit], sv.Length))
		return
	}
	w.writeString(fmt.Sprintf("slice[0x%X]", vbase.Data[sv.Offset:sv.Offset+sv.Length]))
}

func (pv PointerValue) String() string {
	w := newBoundedBuilder()
	pv.writePointer(w, newSeenValues())
	return w.String()
}

// writePointer renders the pointer's "&..." form into w.
func (pv PointerValue) writePointer(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if i := seen.IndexOf(pv); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}
	if !seen.Put(pv) {
		w.writeString("...")
		return
	}
	defer seen.Pop()

	// Handle nil TV's, avoiding a nil pointer deref below.
	if pv.TV == nil {
		w.writeString("&<nil>")
		return
	}
	w.writeByte('&')
	pv.TV.writeWrapped(w, seen)
}

func (sv *StructValue) String() string {
	w := newBoundedBuilder()
	sv.writeBare(w, newSeenValues())
	return w.String()
}

// writeBare renders the struct's "struct{...}" form into w.
func (sv *StructValue) writeBare(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if i := seen.IndexOf(sv); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}
	if !seen.Put(sv) {
		w.writeString("...")
		return
	}
	defer seen.Pop()

	w.writeString("struct{")
	for i := range sv.Fields {
		if w.writeSep(i) {
			break
		}
		sv.Fields[i].writeWrapped(w, seen)
	}
	w.writeByte('}')
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
	w := newBoundedBuilder()
	mv.writeBare(w, newSeenValues())
	return w.String()
}

// writeBare renders the map's "map{...}" form into w.
func (mv *MapValue) writeBare(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if mv.List == nil {
		w.writeString("zero-map")
		return
	}
	if i := seen.IndexOf(mv); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}
	if !seen.Put(mv) {
		w.writeString("...")
		return
	}
	defer seen.Pop()

	if mv.GetLength() > printLimit {
		w.writeString(fmt.Sprintf("map{...(%d entries)}", mv.GetLength()))
		return
	}
	w.writeString("map{")
	i := 0
	for next := mv.List.Head; next != nil; next = next.Next {
		if w.writeSep(i) {
			break
		}
		i++
		next.Key.writeWrapped(w, seen)
		w.writeByte(':')
		next.Value.writeWrapped(w, seen)
	}
	w.writeByte('}')
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
// *TypedValue.Sprint / String

// Sprint is for print() and println().
func (tv *TypedValue) Sprint(m *Machine) string {
	// if undefined, just "undefined".
	if tv == nil || tv.T == nil {
		return undefinedStr
	}

	// if implements .String(), return it.
	if IsImplementedBy(gStringerType, tv.T) && !tv.IsNilInterface() {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "String")))
		return res[0].GetString()
	}
	// if implements .Error(), return it.
	if IsImplementedBy(gErrorType, tv.T) {
		res := m.Eval(Call(Sel(&ConstExpr{TypedValue: *tv}, "Error")))
		return res[0].GetString()
	}

	w := newBoundedBuilder()
	tv.writeSprint(w, newSeenValues(), true)
	return w.String()
}

// String is for gno debugging/testing.
func (tv TypedValue) String() string {
	w := newBoundedBuilder()
	(&tv).writeWrapped(w, newSeenValues())
	return w.String()
}

// writeWrapped renders tv in the "(value type)" form used by String() and for
// every nested element. writeWrapped and writeSprint are mutually recursive and
// both write into w, which enforces the global output cap.
func (tv *TypedValue) writeWrapped(w *boundedBuilder, seen *seenValues) {
	if w.done() {
		return
	}
	if tv.IsUndefined() {
		w.writeString("(undefined)")
		return
	}
	w.writeByte('(')
	if tv.V == nil {
		switch baseOf(tv.T) {
		case BoolType, UntypedBoolType:
			w.writeString(fmt.Sprintf("%t", tv.GetBool()))
		case StringType, UntypedStringType:
			w.writeString(tv.GetString())
		case IntType:
			w.writeString(fmt.Sprintf("%d", tv.GetInt()))
		case Int8Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt8()))
		case Int16Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt16()))
		case Int32Type, UntypedRuneType:
			w.writeString(fmt.Sprintf("%d", tv.GetInt32()))
		case Int64Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt64()))
		case UintType:
			w.writeString(fmt.Sprintf("%d", tv.GetUint()))
		case Uint8Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint8()))
		case DataByteType:
			w.writeString(fmt.Sprintf("%d", tv.GetDataByte()))
		case Uint16Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint16()))
		case Uint32Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint32()))
		case Uint64Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint64()))
		case Float32Type:
			w.writeString(fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32())))
		case Float64Type:
			w.writeString(fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64())))
		// Complex types that require recursion protection.
		default:
			w.writeString(nilStr)
		}
	} else if base := baseOf(tv.T); base == StringType || base == UntypedStringType {
		// Equivalent to quoting the unwrapped (Sprint) form: for a string
		// the Sprint form is exactly tv.GetString() (the seen-index check
		// never matches a StringValue, which is never recorded in seen).
		w.writeString(strconv.Quote(tv.GetString()))
	} else {
		tv.writeSprint(w, seen, false)
	}
	w.writeByte(' ')
	w.writeString(tv.T.String())
	w.writeByte(')')
}

// writeSprint renders tv in the raw print/println form. considerDeclaredType
// routes declared types to the wrapped form; it is true at the top-level Sprint
// entry point and false for values reached recursively.
func (tv *TypedValue) writeSprint(w *boundedBuilder, seen *seenValues, considerDeclaredType bool) {
	if w.done() {
		return
	}
	if i := seen.IndexOf(tv.V); i != -1 {
		w.writeString(fmt.Sprintf("ref@%d", i))
		return
	}

	// print declared type
	if _, ok := tv.T.(*DeclaredType); ok && considerDeclaredType {
		tv.writeWrapped(w, seen)
		return
	}

	// This is a special case that became necessary after adding the protected
	// string machinery to reliably prevent recursive print loops.
	if tv.V != nil {
		if v, ok := tv.V.(RefValue); ok {
			w.writeString(v.String())
			return
		}
	}

	// otherwise, default behavior.
	switch bt := baseOf(tv.T).(type) {
	case PrimitiveType:
		switch bt {
		case UntypedBoolType, BoolType:
			w.writeString(fmt.Sprintf("%t", tv.GetBool()))
		case UntypedStringType, StringType:
			w.writeString(tv.GetString())
		case IntType:
			w.writeString(fmt.Sprintf("%d", tv.GetInt()))
		case Int8Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt8()))
		case Int16Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt16()))
		case UntypedRuneType, Int32Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt32()))
		case Int64Type:
			w.writeString(fmt.Sprintf("%d", tv.GetInt64()))
		case UintType:
			w.writeString(fmt.Sprintf("%d", tv.GetUint()))
		case Uint8Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint8()))
		case DataByteType:
			w.writeString(fmt.Sprintf("%d", tv.GetDataByte()))
		case Uint16Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint16()))
		case Uint32Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint32()))
		case Uint64Type:
			w.writeString(fmt.Sprintf("%d", tv.GetUint64()))
		case Float32Type:
			w.writeString(fmt.Sprintf("%v", math.Float32frombits(tv.GetFloat32())))
		case Float64Type:
			w.writeString(fmt.Sprintf("%v", math.Float64frombits(tv.GetFloat64())))
		case UntypedBigintType:
			w.writeString(tv.V.(BigintValue).V.String())
		case UntypedBigdecType:
			w.writeString(tv.V.(BigdecValue).V.String())
		default:
			panic("should not happen")
		}
	case *PointerType:
		if tv.V == nil {
			w.writeString("typed-nil")
			return
		}
		ro := tv.IsReadonly()
		if ro {
			w.writeString("readonly(")
		}
		tv.V.(PointerValue).writePointer(w, seen)
		if ro {
			w.writeByte(')')
		}
	case *FuncType:
		switch fv := tv.V.(type) {
		case nil:
			w.writeString(nilStr + " " + tv.T.String())
		case *FuncValue, *BoundMethodValue:
			w.writeString(fv.String())
		default:
			panic(fmt.Sprintf(
				"unexpected func type %v",
				reflect.TypeOf(tv.V)))
		}
	case *InterfaceType:
		if debug {
			if tv.DebugHasValue() {
				panic("should not happen")
			}
		}
		w.writeString(nilStr)
	case *DeclaredType:
		panic("should not happen")
	case *PackageType:
		w.writeString(tv.V.(*PackageValue).String())
	case *ChanType:
		panic("not yet implemented")
	case *TypeType:
		w.writeString(tv.V.(TypeValue).String())
	default:
		// The remaining types may have a nil value.
		if tv.V == nil {
			w.writeString("(" + nilStr + " " + tv.T.String() + ")")
			return
		}
		// Value may be N_Readonly
		ro := tv.IsReadonly()
		if ro {
			w.writeString("readonly(")
		}
		// *ArrayType, *SliceType, *StructType, *MapType
		switch cv := tv.V.(type) {
		case *ArrayValue:
			cv.writeBare(w, seen)
		case *SliceValue:
			cv.writeBare(w, seen)
		case *StructValue:
			cv.writeBare(w, seen)
		case *MapValue:
			cv.writeBare(w, seen)
		default:
			if s, ok := tv.V.(fmt.Stringer); ok {
				// *NativeType
				w.writeString(s.String())
			} else if debug {
				panic(fmt.Sprintf(
					"unexpected type %s",
					tv.T.String()))
			} else {
				panic("should not happen")
			}
		}
		if ro {
			w.writeByte(')')
		}
	}
}
