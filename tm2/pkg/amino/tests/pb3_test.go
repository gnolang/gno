package tests

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
)

func TestRoundtripBinary2_PrimitivesStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := PrimitivesStruct{
		Int8:        42,
		Int16:       1000,
		Int32:       100000,
		Int32Fixed:  -1,
		Int64:       99999999,
		Int64Fixed:  -12345,
		Int:         777,
		Byte:        0xFF,
		Uint8:       200,
		Uint16:      50000,
		Uint32:      3000000,
		Uint32Fixed: 42,
		Uint64:      9999999999,
		Uint64Fixed: 1234567890,
		Uint:        12345,
		Str:         "hello",
		Bytes:       []byte{1, 2, 3},
		Time:        time.Unix(1234567890, 123000000).UTC(),
		Duration:    time.Duration(5*time.Second + 500*time.Millisecond),
	}

	compareEncoding(t, cdc, "PrimitivesStruct", orig)
}

func TestRoundtripBinary2_EmptyStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "EmptyStruct", EmptyStruct{})
}

func TestRoundtripBinary2_SlicesStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := SlicesStruct{
		Int8Sl:    []int8{1, -1, 127},
		Int16Sl:   []int16{256, -256},
		Int32Sl:   []int32{100000, -100000},
		Int64Sl:   []int64{999999999, -999999999},
		IntSl:     []int{1, 2, 3},
		ByteSl:    []byte{0xDE, 0xAD},
		Uint8Sl:   []uint8{1, 2, 3},
		Uint16Sl:  []uint16{1000, 2000},
		Uint32Sl:  []uint32{100000, 200000},
		Uint64Sl:  []uint64{1000000, 2000000},
		UintSl:    []uint{1, 2, 3},
		StrSl:     []string{"hello", "world"},
		BytesSl:   [][]byte{{1, 2}, {3, 4}},
		TimeSl:    []time.Time{time.Unix(1000, 0).UTC(), time.Unix(2000, 0).UTC()},
		DurationSl: []time.Duration{time.Second, time.Minute},
		EmptySl:   []EmptyStruct{{}, {}},
	}

	compareEncoding(t, cdc, "SlicesStruct", orig)
}

func TestRoundtripBinary2_PointersStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	i8 := int8(42)
	i16 := int16(1000)
	str := "hello"
	orig := PointersStruct{
		Int8Pt:  &i8,
		Int16Pt: &i16,
		StrPt:   &str,
	}

	compareEncoding(t, cdc, "PointersStruct", orig)
}

func TestRoundtripBinary2_ComplexSt(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := ComplexSt{
		PrField: PrimitivesStruct{
			Int8: 1,
			Str:  "nested",
		},
	}

	compareEncoding(t, cdc, "ComplexSt", orig)
}

func TestRoundtripBinary2_EmbeddedSt1(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := EmbeddedSt1{
		PrimitivesStruct: PrimitivesStruct{
			Int8: 42,
			Str:  "embedded",
		},
	}

	compareEncoding(t, cdc, "EmbeddedSt1", orig)
}

func TestRoundtripBinary2_AminoMarshalerStruct1(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerStruct1{A: 10, B: 20}
	compareEncoding(t, cdc, "AminoMarshalerStruct1", orig)
}

func TestRoundtripBinary2_AminoMarshalerStruct3(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerStruct3{A: 42}
	compareEncoding(t, cdc, "AminoMarshalerStruct3", orig)
}

func TestRoundtripBinary2_AminoMarshalerInt5(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerInt5(42)
	compareEncoding(t, cdc, "AminoMarshalerInt5", orig)
}

func TestRoundtripBinary2_ArraysStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := ArraysStruct{
		Int8Ar:     [4]int8{1, 2, 3, 4},
		Int16Ar:    [4]int16{10, 20, 30, 40},
		Int32Ar:    [4]int32{100, 200, 300, 400},
		IntAr:      [4]int{1, 2, 3, 4},
		ByteAr:     [4]byte{0xDE, 0xAD, 0xBE, 0xEF},
		Uint8Ar:    [4]uint8{1, 2, 3, 4},
		Uint16Ar:   [4]uint16{10, 20, 30, 40},
		Uint32Ar:   [4]uint32{100, 200, 300, 400},
		Uint64Ar:   [4]uint64{1000, 2000, 3000, 4000},
		UintAr:     [4]uint{1, 2, 3, 4},
		StrAr:      [4]string{"a", "b", "c", "d"},
		BytesAr:    [4][]byte{{1}, {2}, {3}, {4}},
		TimeAr:     [4]time.Time{time.Unix(1, 0).UTC(), time.Unix(2, 0).UTC(), time.Unix(3, 0).UTC(), time.Unix(4, 0).UTC()},
		DurationAr: [4]time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second},
		EmptyAr:    [4]EmptyStruct{{}, {}, {}, {}},
	}

	compareEncoding(t, cdc, "ArraysStruct", orig)
}

func TestRoundtripBinary2_ShortArraysStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := ShortArraysStruct{
		TimeAr:     [0]time.Time{},
		DurationAr: [0]time.Duration{},
	}
	compareEncoding(t, cdc, "ShortArraysStruct", orig)
}

func TestRoundtripBinary2_ArraysArraysStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := ArraysArraysStruct{
		Int8ArAr:  [2][2]int8{{1, 2}, {3, 4}},
		Int16ArAr: [2][2]int16{{10, 20}, {30, 40}},
		Int32ArAr: [2][2]int32{{100, 200}, {300, 400}},
		IntArAr:   [2][2]int{{1, 2}, {3, 4}},
		ByteArAr:  [2][2]byte{{0xDE, 0xAD}, {0xBE, 0xEF}},
		Uint8ArAr: [2][2]uint8{{1, 2}, {3, 4}},
		StrArAr:   [2][2]string{{"a", "b"}, {"c", "d"}},
	}
	compareEncoding(t, cdc, "ArraysArraysStruct", orig)
}

func TestRoundtripBinary2_SlicesSlicesStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := SlicesSlicesStruct{
		Int8SlSl:  [][]int8{{1, 2}, {3, 4}},
		Int16SlSl: [][]int16{{10, 20}, {30, 40}},
		Int32SlSl: [][]int32{{100, 200}, {300, 400}},
		IntSlSl:   [][]int{{1, 2}, {3, 4}},
		ByteSlSl:  [][]byte{{0xDE, 0xAD}, {0xBE, 0xEF}},
		Uint8SlSl: [][]uint8{{1, 2}, {3, 4}},
		StrSlSl:   [][]string{{"a", "b"}, {"c", "d"}},
	}
	compareEncoding(t, cdc, "SlicesSlicesStruct", orig)
}

func TestRoundtripBinary2_PointerSlicesStruct(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	i8a, i8b := int8(1), int8(2)
	i16a, i16b := int16(100), int16(200)
	i32a, i32b := int32(1000), int32(2000)
	i64a, i64b := int64(10000), int64(20000)
	i64fa, i64fb := int64(-1), int64(-2)
	ia, ib := int(42), int(43)
	ba, bb := byte(0xDE), byte(0xAD)
	u8a, u8b := uint8(1), uint8(2)
	u16a, u16b := uint16(100), uint16(200)
	u32a, u32b := uint32(1000), uint32(2000)
	u64a, u64b := uint64(10000), uint64(20000)
	ua, ub := uint(42), uint(43)
	sa, sb2 := "hello", "world"
	orig := PointerSlicesStruct{
		Int8PtSl:       []*int8{&i8a, &i8b},
		Int16PtSl:      []*int16{&i16a, &i16b},
		Int32PtSl:      []*int32{&i32a, &i32b},
		Int64PtSl:      []*int64{&i64a, &i64b},
		Int64FixedPtSl: []*int64{&i64fa, &i64fb},
		IntPtSl:        []*int{&ia, &ib},
		BytePtSl:       []*byte{&ba, &bb},
		Uint8PtSl:      []*uint8{&u8a, &u8b},
		Uint16PtSl:     []*uint16{&u16a, &u16b},
		Uint32PtSl:     []*uint32{&u32a, &u32b},
		Uint64PtSl:     []*uint64{&u64a, &u64b},
		UintPtSl:       []*uint{&ua, &ub},
		StrPtSl:        []*string{&sa, &sb2},
	}
	compareEncoding(t, cdc, "PointerSlicesStruct", orig)
}

func TestRoundtripBinary2_EmbeddedSt2(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := EmbeddedSt2{
		PrimitivesStruct: PrimitivesStruct{Int8: 1, Str: "a"},
		ArraysStruct: ArraysStruct{
			Int8Ar:  [4]int8{1, 2, 3, 4},
			ByteAr:  [4]byte{5, 6, 7, 8},
			StrAr:   [4]string{"a", "b", "c", "d"},
			EmptyAr: [4]EmptyStruct{{}, {}, {}, {}},
		},
		SlicesStruct: SlicesStruct{
			Int8Sl: []int8{1, 2},
			StrSl:  []string{"hello"},
		},
		PointersStruct: PointersStruct{},
	}
	compareEncoding(t, cdc, "EmbeddedSt2", orig)
}

func TestRoundtripBinary2_AminoMarshalerStruct2(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerStruct2{
		B: 42,
	}
	compareEncoding(t, cdc, "AminoMarshalerStruct2", orig)
}

func TestRoundtripBinary2_AminoMarshalerInt4(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerInt4(100)
	compareEncoding(t, cdc, "AminoMarshalerInt4", orig)
}

func TestRoundtripBinary2_AminoMarshalerStruct6(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerStruct6{A: 10, B: 20}
	compareEncoding(t, cdc, "AminoMarshalerStruct6", orig)
}

func TestRoundtripBinary2_AminoMarshalerStruct7(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := AminoMarshalerStruct7{A: 5}
	compareEncoding(t, cdc, "AminoMarshalerStruct7", orig)
}

func TestRoundtripBinary2_EmbeddedSt3(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := EmbeddedSt3{
		PrimitivesStruct: &PrimitivesStruct{Int8: 1, Str: "a"},
		EmptyStruct:      &EmptyStruct{},
	}
	compareEncoding(t, cdc, "EmbeddedSt3", orig)
}

func TestRoundtripBinary2_EmbeddedSt4(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := EmbeddedSt4{
		Foo1:             42,
		PrimitivesStruct: PrimitivesStruct{Int8: 1, Str: "test"},
		Foo2:             "hello",
		Foo3:             []byte{1, 2, 3},
		Foo4:             true,
		Foo5:             99,
	}
	compareEncoding(t, cdc, "EmbeddedSt4", orig)
}

func TestRoundtripBinary2_EmbeddedSt5(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()

	orig := EmbeddedSt5{
		Foo1:             42,
		PrimitivesStruct: &PrimitivesStruct{Int8: 1, Str: "test"},
		Foo2:             "hello",
		Foo3:             []byte{1, 2, 3},
		Foo4:             true,
		Foo5:             99,
	}
	compareEncoding(t, cdc, "EmbeddedSt5", orig)
}

// ----------------------------------------
// GnoVM-inspired type tests

func TestRoundtripBinary2_GnoVMPos(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMPos", GnoVMPos{Line: 42, Column: 7})
}

func TestRoundtripBinary2_GnoVMSpan(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMSpan", GnoVMSpan{
		GnoVMPos: GnoVMPos{Line: 10, Column: 3},
		End:      GnoVMPos{Line: 20, Column: 15},
		Num:      5,
	})
}

func TestRoundtripBinary2_GnoVMLocation(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMLocation", GnoVMLocation{
		PkgPath: "gno.land/r/demo",
		File:    "foo.gno",
		GnoVMSpan: GnoVMSpan{
			GnoVMPos: GnoVMPos{Line: 1, Column: 1},
			End:      GnoVMPos{Line: 100, Column: 80},
			Num:      3,
		},
	})
}

func TestRoundtripBinary2_GnoVMAttrs(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMAttrs", GnoVMAttrs{
		GnoVMLocation: GnoVMLocation{
			PkgPath: "gno.land/r/test",
			File:    "bar.gno",
		},
		Label: "myLabel",
		Line:  42,
	})
}

func TestRoundtripBinary2_GnoVMObjectID(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMObjectID", GnoVMObjectID{
		PkgID:   [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		NewTime: 12345678,
	})
}

func TestRoundtripBinary2_GnoVMObjectInfo(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMObjectInfo", GnoVMObjectInfo{
		ID:      GnoVMObjectID{PkgID: [20]byte{0xFF}, NewTime: 99},
		Hash:    [20]byte{0xDE, 0xAD, 0xBE, 0xEF},
		OwnerID: GnoVMObjectID{NewTime: 1},
		ModTime: 42,
	})
}

func TestRoundtripBinary2_GnoVMTypedValue(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMTypedValue", GnoVMTypedValue{
		T: Concrete1{},
		V: Concrete2{},
		N: [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
	})
}

func TestRoundtripBinary2_GnoVMBlock(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMBlock", GnoVMBlock{
		GnoVMObjectInfo: GnoVMObjectInfo{
			ID:      GnoVMObjectID{NewTime: 1},
			ModTime: 5,
		},
		Source: Concrete1{},
		Values: []GnoVMTypedValue{
			{T: Concrete1{}, N: [8]byte{0xAA}},
			{V: Concrete2{}, N: [8]byte{0xBB}},
		},
	})
}

func TestRoundtripBinary2_GnoVMRefValue(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMRefValue", GnoVMRefValue{
		ObjectID: GnoVMObjectID{NewTime: 42},
		Escaped:  true,
		PkgPath:  "gno.land/r/demo",
		Hash:     [20]byte{0x01, 0x02},
	})
}

func TestRoundtripBinary2_GnoVMFieldType(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMFieldType", GnoVMFieldType{
		Name:     "MyField",
		Type:     Concrete1{},
		Embedded: true,
		Tag:      `json:"myfield"`,
	})
}

func TestRoundtripBinary2_GnoVMStructType(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMStructType", GnoVMStructType{
		PkgPath: "gno.land/r/demo",
		Fields: []GnoVMFieldType{
			{Name: "A", Type: Concrete1{}, Tag: `json:"a"`},
			{Name: "B", Type: Concrete2{}, Embedded: true},
		},
	})
}

func TestRoundtripBinary2_GnoVMFuncValue(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMFuncValue", GnoVMFuncValue{
		GnoVMObjectInfo: GnoVMObjectInfo{
			ID:      GnoVMObjectID{NewTime: 3},
			ModTime: 10,
		},
		Type:      Concrete1{},
		IsMethod:  true,
		IsClosure: false,
		Name:      "myFunc",
		Parent:    Concrete2{},
		Captures: []GnoVMTypedValue{
			{T: Concrete1{}, N: [8]byte{0x01}},
		},
		PkgPath: "gno.land/r/demo",
	})
}

func TestRoundtripBinary2_GnoVMDeclaredType(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMDeclaredType", GnoVMDeclaredType{
		PkgPath: "gno.land/r/demo",
		Name:    "MyType",
		ParentLoc: GnoVMLocation{
			PkgPath: "gno.land/r/base",
			File:    "types.gno",
		},
		Base: Concrete1{},
		Methods: []GnoVMTypedValue{
			{T: Concrete2{}, N: [8]byte{0xFF}},
		},
	})
}

func TestRoundtripBinary2_GnoVMNode(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMNode", GnoVMNode{
		GnoVMAttrs: GnoVMAttrs{
			GnoVMLocation: GnoVMLocation{
				PkgPath: "gno.land/r/demo",
				File:    "main.gno",
				GnoVMSpan: GnoVMSpan{
					GnoVMPos: GnoVMPos{Line: 5, Column: 10},
					End:      GnoVMPos{Line: 5, Column: 25},
				},
			},
			Label: "add",
			Line:  5,
		},
		Op:    1,
		Left:  Concrete1{},
		Right: Concrete2{},
		Args:  []Interface1{Concrete1{}, Concrete2{}},
	})
}

func TestRoundtripBinary2_GnoVMFileNode(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMFileNode", GnoVMFileNode{
		GnoVMAttrs: GnoVMAttrs{
			GnoVMLocation: GnoVMLocation{
				PkgPath: "gno.land/r/demo",
				File:    "main.gno",
			},
		},
		FileName: "main.gno",
		PkgName:  "demo",
		Decls:    []Interface1{Concrete1{}, Concrete2{}},
	})
}

func TestRoundtripBinary2_GnoVMPointerValue(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMPointerValue", GnoVMPointerValue{
		TV:    &GnoVMTypedValue{T: Concrete1{}, N: [8]byte{0x42}},
		Base:  Concrete2{},
		Index: 7,
	})
}

func TestRoundtripBinary2_GnoVMSliceValue(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMSliceValue", GnoVMSliceValue{
		Base:   Concrete1{},
		Offset: 10,
		Length: 20,
		Maxcap: 32,
	})
}

func TestRoundtripBinary2_GnoVMMapEntry(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(Package)
	cdc.Seal()
	compareEncoding(t, cdc, "GnoVMMapEntry", GnoVMMapEntry{
		Key:   GnoVMTypedValue{T: Concrete1{}, N: [8]byte{0x01}},
		Value: GnoVMTypedValue{V: Concrete2{}, N: [8]byte{0x02}},
	})
}

// compareEncoding verifies that MarshalBinary2 produces identical bytes to amino.MarshalReflect
// and that UnmarshalBinary2 can roundtrip the data.
func compareEncoding(t *testing.T, cdc *amino.Codec, name string, orig interface{}) {
	t.Helper()

	// Encode with amino reflection.
	bz1, err := cdc.MarshalReflect(orig)
	if err != nil {
		t.Fatalf("%s: MarshalReflect: %v", name, err)
	}

	// Encode with generated MarshalBinary2.
	msg, ok := orig.(amino.PBMessager2)
	if !ok {
		// Try pointer.
		rv := reflect.New(reflect.TypeOf(orig))
		rv.Elem().Set(reflect.ValueOf(orig))
		msg, ok = rv.Interface().(amino.PBMessager2)
		if !ok {
			t.Fatalf("%s: does not implement PBMessager2", name)
		}
	}

	var buf bytes.Buffer
	if err := msg.MarshalBinary2(cdc, &buf); err != nil {
		t.Fatalf("%s: MarshalBinary2: %v", name, err)
	}
	bz2 := buf.Bytes()

	// Check SizeBinary2 matches actual marshal length.
	sizeResult := msg.SizeBinary2(cdc)
	if sizeResult != len(bz2) {
		t.Errorf("%s: SizeBinary2=%d but MarshalBinary2 produced %d bytes", name, sizeResult, len(bz2))
	}

	// Compare.
	if !bytes.Equal(bz1, bz2) {
		t.Errorf("%s: bytes mismatch:\n  amino:     %X\n  genproto2: %X", name, bz1, bz2)
		return
	}

	// Roundtrip unmarshal.
	rt := reflect.TypeOf(orig)
	decoded := reflect.New(rt)
	umsg := decoded.Interface().(amino.PBMessager2)
	if err := umsg.UnmarshalBinary2(cdc, bz1); err != nil {
		t.Fatalf("%s: UnmarshalBinary2: %v", name, err)
	}

	// Re-encode the decoded value and compare.
	var buf2 bytes.Buffer
	// Get the value (not pointer) for marshal.
	decodedVal := decoded.Elem().Interface()
	msg2, ok := decodedVal.(amino.PBMessager2)
	if !ok {
		rv := reflect.New(rt)
		rv.Elem().Set(reflect.ValueOf(decodedVal))
		msg2 = rv.Interface().(amino.PBMessager2)
	}
	if err := msg2.MarshalBinary2(cdc, &buf2); err != nil {
		t.Fatalf("%s: MarshalBinary2 after unmarshal: %v", name, err)
	}
	if !bytes.Equal(bz1, buf2.Bytes()) {
		t.Errorf("%s: roundtrip bytes mismatch:\n  original:  %X\n  roundtrip: %X", name, bz1, buf2.Bytes())
	}
}
