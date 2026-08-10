package gnolang

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// ErrExportSizeExceeded is returned by ExportValues / ExportObject when the
// estimated serialized size of the value tree exceeds the caller's budget.
// The chain's value-returning query endpoints surface this to the client
// instead of walking and then serializing a multi-gigabyte structure; see
// gno.land/adr/prxxxx_query_export_size_guard.md.
var ErrExportSizeExceeded = errors.New("export size limit exceeded")

// exportNodeEst is charged once per exported value/type node to account for
// amino's per-node JSON overhead (braces, field names, commas, the T/V/N
// envelope of a TypedValue). It is a deliberately rough estimate: the guard's
// job is to hard-cap the export walk's output, not to predict the exact JSON
// size.
//
// Every variable-length string the walk *emits* must be charged on top of this,
// at its own length: not just value content (StringValue, []byte Data) but also
// the node-attached names the walk copies out of types — field names, struct
// tags, pkg paths, type IDs. Those are attacker-controlled (a deployed package
// may declare a struct tag of arbitrary length) and, for a non-declared struct
// type, are re-emitted once per element of a slice of that type. Charging only
// per node there made the effective bound nodeCount × maxTagLen: measured, a
// 10KB tag on an anonymous struct passed a 10MB budget while marshaling to
// 198MB. See TestExportValuesLimit_FieldNamesAndTagsCharged.
const exportNodeEst = 32

// exportLimiter bounds the estimated serialized size of an export walk. It is
// threaded through the walk and charged at every visited node; when the running
// estimate exceeds max it panics with ErrExportSizeExceeded, aborting the walk
// early — before the copy completes and before amino marshals anything. The
// entry points (ExportValues / ExportObject) recover the panic into a clean
// error.
//
// A nil *exportLimiter imposes no bound, which is what maxBytes <= 0 selects.
type exportLimiter struct {
	size int64
	max  int64
}

// add charges n estimated bytes and panics with ErrExportSizeExceeded once the
// running total exceeds the budget. Nil-safe: a nil limiter charges nothing.
func (l *exportLimiter) add(n int64) {
	if l == nil {
		return
	}
	l.size += n
	if l.size > l.max {
		panic(ErrExportSizeExceeded)
	}
}

// ExportRefValue represents a back-reference to an ephemeral Object already
// emitted earlier in the export stream. Unlike RefValue (which uses an
// ObjectID for persisted objects), ExportRefValue uses a synthetic ":N" ID,
// where N is an incrementing counter assigned in the encoder's DFS traversal
// order. The first time an ephemeral Object is visited it is expanded inline
// and assigned N = (count of previously-seen ephemeral Objects) + 1; any
// subsequent visit to the same Object emits ExportRefValue{":N"} instead of
// re-expanding it.
//
// Consumers that need to resolve ":N" back to its inline occurrence must walk
// the exported tree in the same DFS order the encoder uses (source-order
// fields, slice/array indices, MapList insertion order, Block values), count
// each inline ephemeral Object as they encounter it, and look up the Nth one.
// Registered with Amino as "/gno.ExportRefValue".
type ExportRefValue struct {
	ObjectID string `json:"ObjectID"` // ":1", ":2", etc.
}

func (ExportRefValue) assertValue() {}
func (erv ExportRefValue) String() string {
	return fmt.Sprintf("exportref(%s)", erv.ObjectID)
}
func (erv ExportRefValue) VisitAssociated(_ Visitor) (stop bool) { return false }

// GetShallowSize returns the size of an ExportRefValue for alloc-gas
// accounting. Uses the same size as RefValue since they're structurally
// equivalent (one ObjectID string).
func (erv ExportRefValue) GetShallowSize() int64 { return allocRefValue }

// ExportValues exports multiple TypedValues for JSON serialization.
// It walks the value tree and:
//   - Replaces persisted (real) objects with RefValue{ObjectID: ...}
//   - Breaks cycles in ephemeral (unreal) objects with ExportRefValue{ObjectID: ":N"}
//   - Copies all values defensively to prevent accidental mutation
//
// The result is suitable for amino.MarshalJSON() serialization.
//
// The walk aborts with ErrExportSizeExceeded as soon as the estimated
// serialized size of the tree exceeds maxBytes, before the copy completes and
// before anything is marshaled — a crafted value tree must not be able to
// force an unbounded export walk + JSON marshal. maxBytes <= 0 disables the
// bound and is only for callers whose input is trusted; anything reachable
// from an ABCI query must pass a real budget.
func ExportValues(tvs []TypedValue, maxBytes int64) ([]TypedValue, error) {
	return withExportLimit(maxBytes, func(lim *exportLimiter) []TypedValue {
		seen := make(map[Object]int)
		result := make([]TypedValue, len(tvs))
		for i, tv := range tvs {
			result[i] = exportValue(tv, seen, lim)
		}
		return result
	})
}

// ExportObject exports a single Object for JSON serialization.
// The object is expanded inline (depth 0), but nested real objects
// become RefValue references. Ephemeral cycles are broken with
// ExportRefValue{":N"} references.
//
// As with ExportValues, the walk aborts with ErrExportSizeExceeded once the
// estimate exceeds maxBytes; maxBytes <= 0 disables the bound.
func ExportObject(obj Object, maxBytes int64) (Value, error) {
	return withExportLimit(maxBytes, func(lim *exportLimiter) Value {
		seen := make(map[Object]int)
		return exportObjectToValue(obj, seen, lim)
	})
}

// withExportLimit runs an export walk under a size budget: it builds the
// limiter (nil when maxBytes <= 0, i.e. unbounded), and recovers the
// ErrExportSizeExceeded panic that add() raises when the walk overshoots,
// returning it as a clean error. Any other panic is re-raised unchanged.
func withExportLimit[T any](maxBytes int64, walk func(*exportLimiter) T) (result T, err error) {
	var lim *exportLimiter
	if maxBytes > 0 {
		lim = &exportLimiter{max: maxBytes}
	}
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok && errors.Is(e, ErrExportSizeExceeded) {
				var zero T
				result, err = zero, ErrExportSizeExceeded
				return
			}
			// Not the size guard: re-raise. The stack restarts at this
			// frame, which is acceptable — the walk's own panics (unexpected
			// value or type kinds) are bugs, and the query layer's recover
			// still reports them.
			panic(r)
		}
	}()
	return walk(lim), nil
}

// exportValue exports a TypedValue, replacing objects with refs.
func exportValue(tv TypedValue, seen map[Object]int, lim *exportLimiter) TypedValue {
	lim.add(exportNodeEst)
	result := TypedValue{N: tv.N}
	if tv.T != nil {
		result.T = exportRefOrCopyType(tv.T, seen, lim)
	}
	if obj, ok := tv.V.(Object); ok {
		result.V = exportToRefOrCopy(obj, seen, lim)
		return result
	}
	if tv.V != nil {
		result.V = exportCopyValue(tv.V, seen, lim)
	}
	return result
}

// exportObjectToValue exports an Object, expanding it inline.
// Nested real objects become RefValue. Ephemeral cycles are broken.
func exportObjectToValue(obj Object, seen map[Object]int, lim *exportLimiter) Value {
	if obj == nil {
		return nil
	}

	// Unwrap HeapItemValue: if the inner value is an Object (ephemeral case),
	// export the inner object directly. For persisted HeapItemValues, the inner
	// value is a RefValue (not an Object), so this is a no-op.
	if hiv, ok := obj.(*HeapItemValue); ok {
		if innerObj, ok := hiv.Value.V.(Object); ok {
			obj = innerObj
		}
	}

	// Check for cycles
	if id, exists := seen[obj]; exists {
		if obj.GetIsReal() {
			return RefValue{
				ObjectID: obj.GetObjectID(),
			}
		}
		return ExportRefValue{
			ObjectID: ":" + strconv.Itoa(id),
		}
	}

	// Mark seen
	id := len(seen) + 1
	seen[obj] = id

	// Expand inline
	return exportCopyValue(obj, seen, lim)
}

// exportToRefOrCopy converts an Object to a RefValue if it's persisted,
// or copies it inline if it's ephemeral.
// This is analogous to realm.go's toRefValue but handles unreal objects
// by assigning synthetic cycle-breaking IDs instead of panicking.
func exportToRefOrCopy(val Value, seen map[Object]int, lim *exportLimiter) Value {
	if ref, ok := val.(RefValue); ok {
		return ref
	}

	oo, ok := val.(Object)
	if !ok {
		panic("unexpected error converting to ref value")
	}

	// Packages always become refs
	if pv, ok := val.(*PackageValue); ok {
		return RefValue{PkgPath: pv.PkgPath}
	}

	// Real (persisted) objects always become RefValue with their real ObjectID.
	// Their children are already RefValues in the store, so cycles are impossible.
	if oo.GetIsReal() {
		return RefValue{
			ObjectID: oo.GetObjectID(),
			Hash:     oo.GetHash(),
		}
	}

	// Unreal (ephemeral) objects: check for cycles
	if id, exists := seen[oo]; exists {
		return ExportRefValue{
			ObjectID: ":" + strconv.Itoa(id),
		}
	}

	// Not yet seen: assign ID, copy inline
	id := len(seen) + 1
	seen[oo] = id
	return exportCopyValue(oo, seen, lim)
}

// exportCopyValue creates a defensive copy of a Value with refs for objects.
// This mirrors realm.go's copyValueWithRefs but handles unreal objects.
func exportCopyValue(val Value, seen map[Object]int, lim *exportLimiter) Value {
	switch cv := val.(type) {
	case nil:
		return nil
	case StringValue:
		lim.add(int64(len(cv)))
		return cv
	case BigintValue:
		// Amino emits the decimal text. 1 digit ≈ 3.32 bits; ÷3 over-charges
		// conservatively, as in bounded_strings.go.
		if cv.V != nil {
			lim.add(int64(cv.V.BitLen())/3 + 1)
		}
		return cv
	case BigdecValue:
		// Amino emits RatString() (rat form) or a hex-float text (float form),
		// not a small constant: a rat component runs up to ratOverflowBits and a
		// float mantissa up to BigdecFloatPrec, both far past exportNodeEst.
		// Charge the emitted length; 1 digit/hex-nibble ≈ 3.32 bits, so ÷3
		// over-charges, as for BigintValue above.
		switch {
		case cv.V != nil:
			lim.add(int64(cv.V.Num().BitLen()+cv.V.Denom().BitLen())/3 + 1)
		case cv.F != nil:
			lim.add(int64(cv.F.Prec())/3 + 1)
		}
		return cv
	case DataByteValue:
		panic("cannot copy data byte value")
	case PointerValue:
		if cv.Base == nil {
			panic("pointer with nil base")
		}
		return PointerValue{
			Base:  exportToRefOrCopy(cv.Base, seen, lim),
			Index: cv.Index,
		}
	case *ArrayValue:
		if cv.Data == nil {
			// Pre-charge the element count before allocating the backing
			// slice: cv.List length is attacker-controllable (e.g. make([]T,
			// n)), so charging only inside the loop would let the full
			// len(cv.List)*sizeof(TypedValue) allocation happen first. Each
			// element is then charged again as it is walked; the resulting
			// double charge is deliberate headroom, not an accident.
			lim.add(int64(len(cv.List)) * exportNodeEst)
			list := make([]TypedValue, len(cv.List))
			for i, etv := range cv.List {
				list[i] = exportValue(etv, seen, lim)
			}
			return &ArrayValue{
				ObjectInfo: cv.ObjectInfo.Copy(),
				List:       list,
			}
		}
		// []byte-backed array: amino serializes Data as base64 (~4/3× the
		// raw length). Charge that before copying so a large byte array is
		// rejected before cp() duplicates it. The binary object endpoint
		// shares this walk but emits raw bytes, so it is rejected ~25%
		// earlier than its output warrants; see keeper.exportObject.
		lim.add(int64(len(cv.Data)) * 4 / 3)
		return &ArrayValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Data:       cp(cv.Data),
		}
	case *SliceValue:
		return &SliceValue{
			Base:   exportToRefOrCopy(cv.Base, seen, lim),
			Offset: cv.Offset,
			Length: cv.Length,
			Maxcap: cv.Maxcap,
		}
	case *StructValue:
		// No pre-charge here (nor in *Block or exportCopyMethods below): the
		// count comes from a declaration, so it cannot be inflated per query
		// the way an array length can.
		fields := make([]TypedValue, len(cv.Fields))
		for i, ftv := range cv.Fields {
			fields[i] = exportValue(ftv, seen, lim)
		}
		return &StructValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Fields:     fields,
		}
	case *FuncValue:
		source := toRefNode(cv.Source)
		var parent Value
		if cv.Parent != nil {
			parent = exportToRefOrCopy(cv.Parent, seen, lim)
		}
		captures := make([]TypedValue, len(cv.Captures))
		for i, ctv := range cv.Captures {
			captures[i] = exportValue(ctv, seen, lim)
		}
		if cv.nativeBody != nil && cv.NativePkg == "" {
			panic("cannot copy function value with native body when there is no native package")
		}
		// Emitted identifiers, charged for the same reason as field
		// names/tags (see exportCopyFieldsWithRefs).
		lim.add(int64(len(cv.Name) + len(cv.FileName) + len(cv.PkgPath) +
			len(cv.NativePkg) + len(cv.NativeName)))
		ft := exportCopyTypeWithRefs(cv.Type, seen, lim)
		return &FuncValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Type:       ft,
			IsMethod:   cv.IsMethod,
			Source:     source,
			Name:       cv.Name,
			Parent:     parent,
			Captures:   captures,
			FileName:   cv.FileName,
			PkgPath:    cv.PkgPath,
			NativePkg:  cv.NativePkg,
			NativeName: cv.NativeName,
			Crossing:   cv.Crossing,
		}
	case *BoundMethodValue:
		// cv.Func is typed *FuncValue, not Value, so it can't carry a
		// RefValue/ExportRefValue back-reference. This mirrors realm.go's
		// copyValueWithRefs pattern. Safe because a BoundMethodValue holds
		// a unique, freshly-constructed FuncValue instance that is not
		// shared with any other traversal path (BoundMethodValue is
		// created at bind time, not deduplicated). If that invariant ever
		// changes, this branch would re-expand a shared FuncValue inline
		// instead of emitting an ExportRefValue — the exported output
		// would still be correct, just potentially larger.
		var fnc *FuncValue // nil for a lazy interface bind (resolved at call)
		if cv.Func != nil {
			fnc = exportCopyValue(cv.Func, seen, lim).(*FuncValue)
		}
		rtv := exportValue(cv.Receiver, seen, lim)
		// Method / MethodPkg are emitted verbatim per node, and for a lazy
		// interface bind (Func == nil) the *FuncValue name charge above never
		// runs — so charge them here for the same reason field names/tags are
		// charged in exportCopyFieldsWithRefs. Method is an attacker-controlled
		// identifier re-emitted once per element of a slice of method values;
		// without this a long method name bypasses the budget (see
		// TestExportValuesLimit_BoundMethodNameCharged).
		lim.add(int64(len(cv.Method) + len(cv.MethodPkg)))
		return &BoundMethodValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Func:       fnc,
			Receiver:   rtv,
			Method:     cv.Method,
			MethodPkg:  cv.MethodPkg,
		}
	case *MapValue:
		list := &MapList{}
		for cur := cv.List.Head; cur != nil; cur = cur.Next {
			key2 := exportValue(cur.Key, seen, lim)
			val2 := exportValue(cur.Value, seen, lim)
			list.Append(nil, key2).Value = val2
		}
		return &MapValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			List:       list,
		}
	case TypeValue:
		// Export the type as a reference, not inline. Consumers that
		// need the full definition (e.g. struct field names, method set)
		// resolve the TypeID via the vm/qtype_json query endpoint.
		// Keeping this symmetric with field-position types (which also
		// go through exportRefOrCopyType at Layer 1) gives a uniform
		// wire shape and smaller JSON payloads.
		return toTypeValue(exportRefOrCopyType(cv.Type, seen, lim))
	case *PackageValue:
		return RefValue{PkgPath: cv.PkgPath}
	case *Block:
		source := toRefNode(cv.Source)
		vals := make([]TypedValue, len(cv.Values))
		for i, tv := range cv.Values {
			vals[i] = exportValue(tv, seen, lim)
		}
		var bparent Value
		if cv.Parent != nil {
			bparent = exportToRefOrCopy(cv.Parent, seen, lim)
		}
		return &Block{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Source:     source,
			Values:     vals,
			Parent:     bparent,
			Blank:      TypedValue{},
		}
	case RefValue:
		return cv
	case *HeapItemValue:
		return &HeapItemValue{
			ObjectInfo: cv.ObjectInfo.Copy(),
			Value:      exportValue(cv.Value, seen, lim),
		}
	default:
		panic(fmt.Sprintf("unexpected type %v", reflect.TypeOf(val)))
	}
}

// exportRefOrCopyType replaces DeclaredTypes with RefType, copies others.
func exportRefOrCopyType(typ Type, seen map[Object]int, lim *exportLimiter) Type {
	if dt, ok := typ.(*DeclaredType); ok {
		// Computed before charging: TypeID's length is bounded by the
		// declaration, so it cannot be the thing that blows the budget.
		id := dt.TypeID()
		// The RefType node and its ID are both emitted here; this branch
		// never reaches exportCopyTypeWithRefs, so charge them locally.
		lim.add(exportNodeEst + int64(len(id)))
		return RefType{ID: id}
	}
	return exportCopyTypeWithRefs(typ, seen, lim)
}

// exportCopyTypeWithRefs copies a type, replacing DeclaredTypes with RefType.
func exportCopyTypeWithRefs(typ Type, seen map[Object]int, lim *exportLimiter) Type {
	lim.add(exportNodeEst)
	switch ct := typ.(type) {
	case nil:
		panic("cannot copy nil types")
	case PrimitiveType:
		return ct
	case *PointerType:
		return &PointerType{
			Elt: exportRefOrCopyType(ct.Elt, seen, lim),
		}
	case FieldType:
		panic("cannot copy field types")
	case *ArrayType:
		return &ArrayType{
			Len: ct.Len,
			Elt: exportRefOrCopyType(ct.Elt, seen, lim),
			Vrd: ct.Vrd,
		}
	case *SliceType:
		return &SliceType{
			Elt: exportRefOrCopyType(ct.Elt, seen, lim),
			Vrd: ct.Vrd,
		}
	case *StructType:
		lim.add(int64(len(ct.PkgPath)))
		return &StructType{
			PkgPath: ct.PkgPath,
			Fields:  exportCopyFieldsWithRefs(ct.Fields, seen, lim),
		}
	case *FuncType:
		return &FuncType{
			Params:  exportCopyFieldsWithRefs(ct.Params, seen, lim),
			Results: exportCopyFieldsWithRefs(ct.Results, seen, lim),
		}
	case *MapType:
		return &MapType{
			Key:   exportRefOrCopyType(ct.Key, seen, lim),
			Value: exportRefOrCopyType(ct.Value, seen, lim),
		}
	case *InterfaceType:
		lim.add(int64(len(ct.PkgPath) + len(ct.Generic)))
		return &InterfaceType{
			PkgPath: ct.PkgPath,
			Methods: exportCopyFieldsWithRefs(ct.Methods, seen, lim),
			Generic: ct.Generic,
		}
	case *TypeType:
		return &TypeType{}
	case *DeclaredType:
		// Likely dead code. Every path that could hand a *DeclaredType
		// to this function now routes through exportRefOrCopyType at
		// Layer 1 instead, which collapses DeclaredTypes to RefType{ID}
		// before reaching this switch: field/element types via
		// exportCopyFieldsWithRefs, TypeValue positions in
		// exportCopyValue, and tv.T in exportValue. *FuncValue.Type and
		// method mtv.T are *FuncType, never DeclaredType.
		// DeclaredType.Base is invariantly non-DeclaredType per the
		// types.go:1441 doc comment (enforced by declareWith/baseOf).
		// Kept as defensive code; if a future caller hands in a
		// DeclaredType directly, the inlined form here + exportCopyMethods
		// below both still produce correct output.
		lim.add(int64(len(ct.PkgPath) + len(ct.Name)))
		dt := &DeclaredType{
			PkgPath: ct.PkgPath,
			Name:    ct.Name,
			Base:    exportCopyTypeWithRefs(ct.Base, seen, lim),
			Methods: exportCopyMethods(ct.Methods, seen, lim),
		}
		return dt
	case *PackageType:
		return &PackageType{}
	case *ChanType:
		return &ChanType{
			Dir: ct.Dir,
			Elt: exportRefOrCopyType(ct.Elt, seen, lim),
		}
	case blockType:
		return blockType{}
	case *tupleType:
		elts2 := make([]Type, len(ct.Elts))
		for i, elt := range ct.Elts {
			elts2[i] = exportRefOrCopyType(elt, seen, lim)
		}
		return &tupleType{
			Elts: elts2,
		}
	case RefType:
		lim.add(int64(len(ct.ID)))
		return RefType{ID: ct.ID}
	case heapItemType:
		return ct
	default:
		panic(fmt.Sprintf("unexpected type %v", typ))
	}
}

func exportCopyFieldsWithRefs(fields []FieldType, seen map[Object]int, lim *exportLimiter) []FieldType {
	// Pre-charge the field count before allocating the backing slice, then
	// charge each field's emitted strings: Name/Tag/PkgPath are copied out
	// verbatim and are attacker-controlled. This is the site of the bypass
	// described in exportNodeEst's doc comment — do not reduce these to a
	// per-node charge.
	lim.add(int64(len(fields)) * exportNodeEst)
	fieldsCpy := make([]FieldType, len(fields))
	for i, field := range fields {
		lim.add(int64(len(field.Name) + len(field.Tag) + len(field.PkgPath)))
		fieldsCpy[i] = FieldType{
			Name:     field.Name,
			Type:     exportRefOrCopyType(field.Type, seen, lim),
			Embedded: field.Embedded,
			Tag:      field.Tag,
			PkgPath:  field.PkgPath,
		}
	}
	return fieldsCpy
}

// exportCopyMethods is reached only from the *DeclaredType branch of
// exportCopyTypeWithRefs, which is itself likely dead code post-fix
// (see comment there). Kept as defensive code. One caveat if it does
// ever fire: V is expanded via exportCopyValue rather than
// exportToRefOrCopy, so if the same *FuncValue is reachable elsewhere
// in the exported tree (e.g. via a BoundMethodValue holding it), it
// gets re-expanded inline rather than deduplicated. Acceptable because
// the inlined copies are byte-identical; a consumer sees duplication
// but not inconsistency.
func exportCopyMethods(methods []TypedValue, seen map[Object]int, lim *exportLimiter) []TypedValue {
	res := make([]TypedValue, len(methods))
	for i, mtv := range methods {
		res[i] = TypedValue{
			T: exportCopyTypeWithRefs(mtv.T, seen, lim),
			V: exportCopyValue(mtv.V, seen, lim),
		}
	}
	return res
}
