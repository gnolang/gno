//go:build genproto2

package benchstore

// Constructs realistic gnolang values for amino benchmarking.
// All child object references are replaced with RefValue{} to match
// what the store persists (see copyValueWithRefs in realm.go).

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

var benchPkgID = gno.PkgIDFromPkgPath("gno.land/r/bench")

// nextOID is a simple counter for generating unique ObjectIDs.
var nextOID uint64

func newOID() gno.ObjectID {
	nextOID++
	return gno.ObjectID{PkgID: benchPkgID, NewTime: nextOID}
}

func refVal() gno.RefValue {
	return gno.RefValue{ObjectID: newOID(), Hash: gno.ValueHash{Hashlet: gno.HashBytes([]byte("x"))}}
}

// typedRef returns a TypedValue whose V is a RefValue (as persisted).
func typedRef(t gno.Type) gno.TypedValue {
	return gno.TypedValue{T: t, V: refVal()}
}

// typedInt returns a TypedValue holding an int primitive.
func typedInt(n int) gno.TypedValue {
	tv := gno.TypedValue{T: gno.IntType}
	tv.SetInt(int64(n))
	return tv
}

// typedString returns a TypedValue holding a short string.
func typedString(s string) gno.TypedValue {
	return gno.TypedValue{T: gno.StringType, V: gno.StringValue(s)}
}

// typedBool returns a TypedValue holding a bool.
func typedBool(b bool) gno.TypedValue {
	tv := gno.TypedValue{T: gno.BoolType}
	tv.SetBool(b)
	return tv
}

// ============================================================
// Value constructors — produce values as they'd appear after
// copyValueWithRefs (child objects replaced with RefValue).
// ============================================================

// makeArrayData makes an ArrayValue with byte data (like []byte).
func makeArrayData(n int) *gno.ArrayValue {
	av := &gno.ArrayValue{Data: make([]byte, n)}
	av.SetObjectID(newOID())
	av.IncRefCount()
	for i := range av.Data {
		av.Data[i] = byte(i)
	}
	return av
}

// makeArrayList makes an ArrayValue with a typed element list.
func makeArrayList(n int) *gno.ArrayValue {
	av := &gno.ArrayValue{List: make([]gno.TypedValue, n)}
	av.SetObjectID(newOID())
	av.IncRefCount()
	for i := range av.List {
		av.List[i] = typedInt(i)
	}
	return av
}

// makeStruct makes a StructValue with n fields of mixed types.
func makeStruct(nFields int) *gno.StructValue {
	sv := &gno.StructValue{Fields: make([]gno.TypedValue, nFields)}
	sv.SetObjectID(newOID())
	sv.IncRefCount()
	for i := range sv.Fields {
		switch i % 4 {
		case 0:
			sv.Fields[i] = typedInt(i * 100)
		case 1:
			sv.Fields[i] = typedString(fmt.Sprintf("f%d", i))
		case 2:
			sv.Fields[i] = typedBool(i%2 == 0)
		case 3:
			// a reference to another object (as persisted)
			sv.Fields[i] = typedRef(gno.IntType)
		}
	}
	return sv
}

// makeFunc makes a FuncValue as it would be persisted.
func makeFunc(nCaptures int) *gno.FuncValue {
	fv := &gno.FuncValue{
		Type: &gno.FuncType{
			Params:  []gno.FieldType{{Name: "a", Type: gno.IntType}, {Name: "b", Type: gno.StringType}},
			Results: []gno.FieldType{{Name: "", Type: gno.IntType}},
		},
		IsMethod: false,
		Source: gno.RefNode{Location: gno.Location{
			PkgPath: "gno.land/r/bench",
			File:    "bench.gno",
		}},
		Name:     "benchFn",
		Parent:   refVal(), // block ref
		Captures: make([]gno.TypedValue, nCaptures),
		FileName: "bench.gno",
		PkgPath:  "gno.land/r/bench",
	}
	fv.SetObjectID(newOID())
	fv.IncRefCount()
	for i := range fv.Captures {
		fv.Captures[i] = typedRef(gno.IntType) // HeapItemValue refs
	}
	return fv
}

// makeBoundMethod makes a BoundMethodValue.
func makeBoundMethod() *gno.BoundMethodValue {
	bmv := &gno.BoundMethodValue{
		Func:     makeFunc(0),
		Receiver: typedRef(&gno.StructType{PkgPath: "gno.land/r/bench", Fields: []gno.FieldType{{Name: "x", Type: gno.IntType}}}),
	}
	bmv.SetObjectID(newOID())
	bmv.IncRefCount()
	return bmv
}

// makeMap makes a MapValue with n string->int entries.
func makeMap(n int) *gno.MapValue {
	mv := &gno.MapValue{List: &gno.MapList{}}
	mv.SetObjectID(newOID())
	mv.IncRefCount()
	cur := mv.List
	for i := 0; i < n; i++ {
		item := &gno.MapListItem{
			Key:   typedString(fmt.Sprintf("k%d", i)),
			Value: typedInt(i),
		}
		if cur.Head == nil {
			cur.Head = item
			cur.Tail = item
		} else {
			item.Prev = cur.Tail
			cur.Tail.Next = item
			cur.Tail = item
		}
		cur.Size++
	}
	return mv
}

// makeSlice makes a SliceValue referencing an array.
func makeSlice(length int) *gno.SliceValue {
	return &gno.SliceValue{
		Base:   refVal(), // array ref
		Offset: 0,
		Length: length,
		Maxcap: length,
	}
}

// makeBlock makes a Block with n values.
func makeBlock(nValues int) *gno.Block {
	b := &gno.Block{
		Source: gno.RefNode{Location: gno.Location{
			PkgPath: "gno.land/r/bench",
			File:    "bench.gno",
		}},
		Values: make([]gno.TypedValue, nValues),
		Parent: refVal(),
	}
	b.SetObjectID(newOID())
	b.IncRefCount()
	for i := range b.Values {
		switch i % 3 {
		case 0:
			b.Values[i] = typedInt(i)
		case 1:
			b.Values[i] = typedString(fmt.Sprintf("v%d", i))
		case 2:
			b.Values[i] = typedRef(gno.IntType)
		}
	}
	return b
}

// makeHeapItem makes a HeapItemValue.
func makeHeapItem() *gno.HeapItemValue {
	hiv := &gno.HeapItemValue{
		Value: typedInt(42),
	}
	hiv.SetObjectID(newOID())
	hiv.IncRefCount()
	return hiv
}

// makePackage makes a PackageValue as persisted.
func makePackage(nFBlocks int) *gno.PackageValue {
	pv := &gno.PackageValue{
		Block:   refVal(),
		PkgName: "bench",
		PkgPath: "gno.land/r/bench",
		FNames:  make([]string, nFBlocks),
		FBlocks: make([]gno.Value, nFBlocks),
	}
	pv.SetObjectID(newOID())
	pv.IncRefCount()
	for i := range pv.FNames {
		pv.FNames[i] = fmt.Sprintf("file%d.gno", i)
		pv.FBlocks[i] = refVal()
	}
	return pv
}

// ============================================================
// Type constructors
// ============================================================

func makeStructType(nFields int) *gno.StructType {
	fields := make([]gno.FieldType, nFields)
	for i := range fields {
		fields[i] = gno.FieldType{
			Name: gno.Name(fmt.Sprintf("F%d", i)),
			Type: gno.IntType,
		}
	}
	return &gno.StructType{PkgPath: "gno.land/r/bench", Fields: fields}
}

func makeFuncType(nParams, nResults int) *gno.FuncType {
	params := make([]gno.FieldType, nParams)
	for i := range params {
		params[i] = gno.FieldType{Name: gno.Name(fmt.Sprintf("p%d", i)), Type: gno.IntType}
	}
	results := make([]gno.FieldType, nResults)
	for i := range results {
		results[i] = gno.FieldType{Name: gno.Name(fmt.Sprintf("r%d", i)), Type: gno.StringType}
	}
	return &gno.FuncType{Params: params, Results: results}
}

func makeInterfaceType(nMethods int) *gno.InterfaceType {
	methods := make([]gno.FieldType, nMethods)
	for i := range methods {
		methods[i] = gno.FieldType{
			Name: gno.Name(fmt.Sprintf("M%d", i)),
			Type: makeFuncType(1, 1),
		}
	}
	return &gno.InterfaceType{PkgPath: "gno.land/r/bench", Methods: methods}
}

// ============================================================
// TestValue wraps a named value with its amino-serialized bytes.
// ============================================================

type TestValue struct {
	Name  string
	Value interface{} // the gno value or type
	Bytes []byte      // amino-serialized
}

// BuildTestValues constructs a diverse set of realistic gnolang values
// at various sizes, serializes them, and returns them sorted by byte size.
func BuildTestValues() []TestValue {
	var tvs []TestValue

	add := func(name string, v interface{}) {
		bz := amino.MustMarshalAny(v)
		tvs = append(tvs, TestValue{Name: name, Value: v, Bytes: bz})
	}

	// ArrayValue with byte data
	for _, n := range []int{0, 8, 32, 128, 512} {
		add(fmt.Sprintf("ArrayData/%d", n), makeArrayData(n))
	}

	// ArrayValue with typed list
	for _, n := range []int{1, 4, 16, 64} {
		add(fmt.Sprintf("ArrayList/%d", n), makeArrayList(n))
	}

	// StructValue
	for _, n := range []int{1, 4, 8, 16, 32} {
		add(fmt.Sprintf("Struct/%dfields", n), makeStruct(n))
	}

	// FuncValue
	for _, n := range []int{0, 2, 8, 16} {
		add(fmt.Sprintf("Func/%dcap", n), makeFunc(n))
	}

	// BoundMethodValue
	add("BoundMethod", makeBoundMethod())

	// MapValue
	for _, n := range []int{1, 4, 16, 64} {
		add(fmt.Sprintf("Map/%dentries", n), makeMap(n))
	}

	// SliceValue
	add("Slice/16", makeSlice(16))
	add("Slice/256", makeSlice(256))

	// Block
	for _, n := range []int{1, 4, 16, 32} {
		add(fmt.Sprintf("Block/%dvals", n), makeBlock(n))
	}

	// HeapItemValue
	add("HeapItem", makeHeapItem())

	// PackageValue
	for _, n := range []int{1, 4, 8} {
		add(fmt.Sprintf("Package/%dfiles", n), makePackage(n))
	}

	// Types (persisted as TypeValue wrappers)
	for _, n := range []int{1, 4, 8, 16} {
		add(fmt.Sprintf("StructType/%dfields", n), gno.TypeValue{Type: makeStructType(n)})
	}
	for _, n := range []int{1, 4, 8} {
		add(fmt.Sprintf("FuncType/%dparams", n), gno.TypeValue{Type: makeFuncType(n, 1)})
	}
	for _, n := range []int{1, 4, 8} {
		add(fmt.Sprintf("InterfaceType/%dmethods", n), gno.TypeValue{Type: makeInterfaceType(n)})
	}

	return tvs
}
