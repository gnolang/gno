package gnolang

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

// ft builds a concrete (non-embedded) method entry.
func mkMethod(name string, typ Type) FieldType {
	return FieldType{Name: Name(name), Type: typ}
}

// mkEmbed builds an embed entry: a field whose type is an (already flattened)
// interface from package pkg with the given methods.
func mkEmbed(pkg string, methods ...FieldType) FieldType {
	return FieldType{Type: &InterfaceType{PkgPath: pkg, Methods: methods}}
}

func TestFlattenInterfaceMethods(t *testing.T) {
	t.Parallel()

	fn := &FuncType{} // func()

	type want struct {
		name string
		pkg  string
	}
	cases := []struct {
		desc    string
		in      []FieldType
		pkgPath string
		want    []want
	}{
		{
			// No embed → fast path returns the slice unchanged: a direct
			// unexported method is NOT stamped (idName falls back to the
			// enclosing pkg at TypeID time), so PkgPath stays empty.
			desc:    "no-embed interface returned unchanged (fast path)",
			in:      []FieldType{mkMethod("M", fn), mkMethod("m", fn)},
			pkgPath: "q",
			want:    []want{{"M", ""}, {"m", ""}},
		},
		{
			// On the slow path a same-package unexported method is still left
			// unstamped (only cross-package methods are stamped), so it relies
			// on idName's fallback — matching the fast path.
			desc: "same-package unexported left unstamped even on the slow path",
			in: []FieldType{
				mkEmbed("p", mkMethod("E", fn)),
				mkMethod("m", fn),
			},
			pkgPath: "q",
			want:    []want{{"E", ""}, {"m", ""}},
		},
		{
			desc:    "embedded exported method flattens, pkg stays empty (package-independent identity)",
			in:      []FieldType{mkEmbed("p", mkMethod("E", fn))},
			pkgPath: "q",
			want:    []want{{"E", ""}},
		},
		{
			desc:    "embedded unexported method keeps its origin package, not the enclosing one",
			in:      []FieldType{mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"})},
			pkgPath: "q",
			want:    []want{{"sec", "p"}},
		},
		{
			desc: "diamond: same unexported method via two embeds dedups to one",
			in: []FieldType{
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
			},
			pkgPath: "q",
			want:    []want{{"sec", "p"}},
		},
		{
			// cross-package sec (stamped "p") and a same-package sec (left
			// unstamped, qualified to "q" via fallback) are distinct and coexist.
			desc: "same name, different package: distinct methods coexist",
			in: []FieldType{
				mkEmbed("p", FieldType{Name: "sec", Type: fn, PkgPath: "p"}),
				mkMethod("sec", fn), // declared directly in enclosing q
			},
			pkgPath: "q",
			want:    []want{{"sec", "p"}, {"sec", ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()
			got := flattenInterfaceMethods(tc.in, tc.pkgPath)
			if len(got) != len(tc.want) {
				t.Fatalf("len = %d, want %d (%+v)", len(got), len(tc.want), got)
			}
			for i, w := range tc.want {
				if string(got[i].Name) != w.name || got[i].PkgPath != w.pkg {
					t.Errorf("entry %d = (%s, %q), want (%s, %q)",
						i, got[i].Name, got[i].PkgPath, w.name, w.pkg)
				}
				if got[i].Embedded {
					t.Errorf("entry %d (%s) should not be marked Embedded", i, got[i].Name)
				}
			}
		})
	}
}

// Two embeds contributing the same method name with conflicting signatures is
// a should-not-happen (go/types rejects it upstream); flatten's dedup guards
// with a panic. (A pure direct-method duplicate takes the fast path and is
// instead caught later by sortForPackage at TypeID time.)
func TestFlattenInterfaceMethods_ConflictPanics(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	fn2 := &FuncType{Results: []FieldType{{Type: BoolType}}}
	in := []FieldType{
		mkEmbed("p", mkMethod("M", fn)),
		mkEmbed("p", mkMethod("M", fn2)),
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on conflicting embedded method, got none")
		}
	}()
	flattenInterfaceMethods(in, "q")
}

// Regression for the sort/emit provenance mismatch (thehowl review). The same
// interface in package p can hold an unexported method with PkgPath either
// stamped to "p" (the method was hoisted from an embed — slow path) or empty
// (the method was declared directly, so the no-embed fast path leaves it
// unstamped). Both must yield one TypeID: the sort key and the emitted id
// both qualify via the enclosing package, so the order can't flip between
// the two representations.
func TestInterfaceTypeID_PkgPathProvenance(t *testing.T) {
	t.Parallel()
	fn := &FuncType{}
	stamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn, PkgPath: "p"}, // hoisted from an embed
	}}
	unstamped := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "M", Type: fn},
		{Name: "z", Type: fn}, // declared directly (fast path, unstamped)
	}}
	if stamped.TypeID() != unstamped.TypeID() {
		t.Fatalf("stamped vs empty PkgPath gave different TypeIDs:\n stamped:   %s\n unstamped: %s",
			stamped.TypeID(), unstamped.TypeID())
	}
}

// An InterfaceKind entry in Methods can only be state persisted before
// interface flattening (every construction path flattens), which is
// unsupported — identity already moved. Store bytes are external input, so
// the decode boundary (fillType, reached from both GetTypeSafe and
// fillTypesOfValue) rejects it unconditionally; the interior sites
// (FindEmbeddedFieldType/checkImplementedBy/TypeID) assume the invariant
// and only assert under -tags debugAssert. Pins the drop of the legacy
// embedded-interface branches; see adr/pr5739.
func TestInterfaceType_UnflattenedIsHardError(t *testing.T) {
	t.Parallel()
	embedded := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "M", Type: &FuncType{}}}}
	legacy := &InterfaceType{PkgPath: "p", Methods: []FieldType{
		{Name: "E", Type: embedded, Embedded: true}, // as decoded from bytes persisted before flattening
	}}

	uses := map[string]func(){
		// no RefType inside, so a nil store never dereferences
		"fillType": func() { fillType(nil, legacy) },
	}
	if debugAssert {
		uses["FindEmbeddedFieldType"] = func() { legacy.FindEmbeddedFieldType("p", "M") }
		uses["checkImplementedBy"] = func() { legacy.checkImplementedBy(nil, embedded) }
		uses["TypeID"] = func() { legacy.TypeID() }
	}
	for name, use := range uses {
		t.Run(name, func(t *testing.T) {
			t.Parallel() // uses is read-only on the shared types; TypeID panics before its cache write
			defer func() {
				r := recover()
				if r == nil {
					t.Fatal("expected panic on unflattened interface, got none")
				}
				if s, ok := r.(string); !ok || !strings.Contains(s, "unflattened embedded interface") {
					t.Fatalf("expected unflattened-interface panic, got: %v", r)
				}
			}()
			use()
		})
	}
}

// The runtime interface-construction path (doOpInterfaceType) must flatten its
// embeds just like the preprocess path. This path executes during filetests
// (e.g. v.(interface{ Embed })) but its result is never observed for identity
// there, so no filetest pins it; drive the op directly and assert the built
// InterfaceType holds the embed's methods, with no embedded-interface entry.
func TestDoOpInterfaceType_Flattens(t *testing.T) {
	m := NewMachine("p", nil)
	defer m.Release()

	// Two embeds that overlap on B (a diamond): {A,B} and {B,C}. Flattening
	// expands both and dedups B, so the result is exactly {A,B,C} — 3 methods.
	// Without flattening the interface would instead hold 2 embedded-interface
	// entries, so the count distinguishes real flatten+dedup from pass-through.
	fnB := &FuncType{} // shared signature so the overlapping B dedups (not conflicts)
	e1 := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "A", Type: &FuncType{}}, {Name: "B", Type: fnB}}}
	e2 := &InterfaceType{PkgPath: "p", Methods: []FieldType{{Name: "B", Type: fnB}, {Name: "C", Type: &FuncType{}}}}

	// interface{ e1; e2 }: doOpInterfaceType reads len(x.Methods) from the expr
	// and pops one resolved type per method off the value stack.
	m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(FieldType{Type: e1})})
	m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(FieldType{Type: e2})})
	m.PushExpr(&InterfaceTypeExpr{Methods: FieldTypeExprs{{}, {}}})

	m.doOpInterfaceType()

	it := m.PopValue().V.(TypeValue).Type.(*InterfaceType)
	if len(it.Methods) != 3 {
		t.Fatalf("expected 3 flattened+deduped methods (A,B,C), got %d: %+v", len(it.Methods), it.Methods)
	}
	got := map[Name]int{}
	for _, ft := range it.Methods {
		if ft.Type.Kind() == InterfaceKind {
			t.Fatalf("embedded-interface entry survived flattening: %+v", ft)
		}
		got[ft.Name]++
	}
	for _, n := range []Name{"A", "B", "C"} {
		if got[n] != 1 {
			t.Fatalf("method %s appears %d times, want exactly 1: %+v", n, got[n], it.Methods)
		}
	}
}

// The runtime struct-construction path (doOpStructType) must name embedded
// fields from the source expr like the preprocess path (buildFieldTypesAST)
// does. Nothing in the Files suite reaches this op with an embedded field
// (verified by instrumentation), so drive the op directly: the popped
// FieldType carries only the resolved type (the alias is gone), and the name
// must come from the expr as written.
func TestDoOpStructType_EmbedNames(t *testing.T) {
	m := NewMachine("p", nil)
	defer m.Release()

	// struct{ Int; *MyInt; pkg.T; N int } — three embed spellings + one named
	// field (must pass through untouched).
	exprs := FieldTypeExprs{
		{Type: Nx("Int")},                     // alias embed: resolved type int, spelled Int
		{Type: &StarExpr{X: Nx("MyInt")}},     // pointer embed: name from elem
		{Type: Sel(Nx("pkg"), "T")},           // qualified embed: name from selector
		{NameExpr: *Nx("N"), Type: Nx("int")}, // named field: not an embed
	}
	fts := []FieldType{
		{Type: IntType},
		{Type: &PointerType{Elt: IntType}},
		{Type: IntType},
		{Name: "N", Type: IntType},
	}
	for _, ft := range fts {
		m.PushValue(TypedValue{T: gTypeType, V: toTypeValue(ft)})
	}
	m.PushExpr(&StructTypeExpr{Fields: exprs})

	m.doOpStructType()

	st := m.PopValue().V.(TypeValue).Type.(*StructType)
	want := []struct {
		name  Name
		embed bool
	}{{"Int", true}, {"MyInt", true}, {"T", true}, {"N", false}}
	for i, w := range want {
		f := st.Fields[i]
		if f.Name != w.name || f.Embedded != w.embed {
			t.Fatalf("field %d: got (name=%s, embedded=%v), want (name=%s, embedded=%v)",
				i, f.Name, f.Embedded, w.name, w.embed)
		}
	}
}

// An id of exactly n bytes with no id builder behind it: a declared type's id
// is its path and name.
func typeIDOfLen(n int) *DeclaredType {
	return &DeclaredType{Name: Name(strings.Repeat("a", n-1)), Base: IntType}
}

// Each builder is given a child sized to reach the cap exactly, then one byte more.
func TestTypeIDRefusesPastTheCap(t *testing.T) {
	type id = func() TypeID
	field := func(name string, typ Type) FieldType { return FieldType{Name: Name(name), Type: typ} }
	for name, tc := range map[string]struct {
		overhead int
		build    func(child Type) id
	}{
		"array":         {3, func(c Type) id { return (&ArrayType{Len: 2, Elt: c}).TypeID }},
		"slice":         {2, func(c Type) id { return (&SliceType{Elt: c}).TypeID }},
		"pointer":       {1, func(c Type) id { return (&PointerType{Elt: c}).TypeID }},
		"map key":       {8, func(c Type) id { return (&MapType{Key: c, Value: IntType}).TypeID }},
		"map value":     {8, func(c Type) id { return (&MapType{Key: IntType, Value: c}).TypeID }},
		"chan":          {6, func(c Type) id { return (&ChanType{Dir: SEND | RECV, Elt: c}).TypeID }},
		"send chan":     {8, func(c Type) id { return (&ChanType{Dir: SEND, Elt: c}).TypeID }},
		"recv chan":     {8, func(c Type) id { return (&ChanType{Dir: RECV, Elt: c}).TypeID }},
		"func param":    {8, func(c Type) id { return (&FuncType{Params: []FieldType{{Type: c}}}).TypeID }},
		"func result":   {8, func(c Type) id { return (&FuncType{Results: []FieldType{{Type: c}}}).TypeID }},
		"struct":        {10, func(c Type) id { return (&StructType{Fields: []FieldType{field("F", c)}}).TypeID }},
		"interface":     {13, func(c Type) id { return (&InterfaceType{Methods: []FieldType{field("M", c)}}).TypeID }},
		"tuple":         {2, func(c Type) id { return (&tupleType{Elts: []Type{c}}).TypeID }},
		"field":         {2, func(c Type) id { return field("F", c).TypeID }},
		"unnamed field": {0, func(c Type) id { return FieldType{Type: c}.TypeID }},
		"list":          {2, func(c Type) id { return FieldTypeList{field("F", c)}.TypeID }},
		"unnamed list": {4, func(c Type) id {
			return func() TypeID { return FieldTypeList{{Type: IntType}, {Type: c}}.unnamedTypeID(false) }
		}},
		"package list": {4, func(c Type) id {
			return func() TypeID { return FieldTypeList{field("f", c)}.typeIDForPackage("x", false) }
		}},
	} {
		t.Run(name, func(t *testing.T) {
			got := tc.build(typeIDOfLen(maxTypeIDBytes - tc.overhead))()
			if len(got) != maxTypeIDBytes {
				t.Errorf("id at the cap: len %d", len(got))
			}
			over := tc.build(typeIDOfLen(maxTypeIDBytes - tc.overhead + 1))
			defer func() {
				if r := recover(); !strings.Contains(fmt.Sprint(r), "type too large") {
					t.Errorf("id past the cap: recovered %v", r)
				}
			}()
			over()
		})
	}
}

// The id below the cap is what the Sprintf-built TypeID methods wrote, kind by kind.
func TestTypeIDForms(t *testing.T) {
	dt := &DeclaredType{PkgPath: "gno.land/p/x", Name: "N", Base: IntType}
	field := func(name string, typ Type) FieldType { return FieldType{Name: Name(name), Type: typ} }
	for want, typ := range map[string]Type{
		"[2]int":                             &ArrayType{Len: 2, Elt: IntType},
		"[]int":                              &SliceType{Elt: IntType, Vrd: true},
		"*gno.land/p/x.N":                    &PointerType{Elt: dt},
		"map[string]int":                     &MapType{Key: StringType, Value: IntType},
		"chan{int}":                          &ChanType{Dir: SEND | RECV, Elt: IntType},
		"<-chan{int}":                        &ChanType{Dir: SEND, Elt: IntType},
		"chan<-{int}":                        &ChanType{Dir: RECV, Elt: IntType},
		"func()()":                           &FuncType{},
		"func(int;string)()":                 &FuncType{Params: []FieldType{{Type: IntType}, {Type: StringType}}},
		"func()(int;string)":                 &FuncType{Results: []FieldType{{Type: IntType}, {Type: StringType}}},
		"struct{}":                           &StructType{},
		"struct{A int;p.b string}":           &StructType{PkgPath: "p", Fields: []FieldType{field("A", IntType), field("b", StringType)}},
		"interface{A func()();p.b func()()}": &InterfaceType{PkgPath: "p", Methods: []FieldType{field("b", &FuncType{}), field("A", &FuncType{})}},
		"(int,string)":                       &tupleType{Elts: []Type{IntType, StringType}},
		"F int":                              field("F", IntType),
		"int":                                FieldType{Type: IntType},
	} {
		if got := typ.TypeID().String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
	list := FieldTypeList{field("F", IntType), {Type: StringType}}
	if got := list.TypeID().String(); got != "F int;string" {
		t.Errorf("list id: %q", got)
	}
}

// A refusal stops at the child that would pass the cap, so what the parent
// wrote before it is all it costs, however many children follow. Each child's id
// is memoised first, which leaves only the parent's writes to count.
func TestTypeIDRefusalCostsOneCap(t *testing.T) {
	fields := make([]FieldType, 1000)
	child := &SliceType{Elt: typeIDOfLen(maxTypeIDBytes - 100)}
	child.TypeID()
	for i := range fields {
		fields[i] = FieldType{Name: "F", Type: child}
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	func() {
		defer func() { _ = recover() }()
		FieldTypeList(fields).unnamedTypeID(false)
	}()
	runtime.ReadMemStats(&after)
	if n := after.TotalAlloc - before.TotalAlloc; n >= 2*maxTypeIDBytes {
		t.Errorf("refusing allocated %d bytes", n)
	}
}

// The text below the cap is what fmt-built String methods wrote, kind by kind.
func TestTypeTextForms(t *testing.T) {
	dt := &DeclaredType{PkgPath: "gno.land/p/x", Name: "N", Base: IntType}
	for want, typ := range map[string]Type{
		"[2]int":                  &ArrayType{Len: 2, Elt: IntType},
		"[]int":                   &SliceType{Elt: IntType},
		"...int":                  &SliceType{Elt: IntType, Vrd: true},
		"*gno.land/p/x.N":         &PointerType{Elt: dt},
		"map[string]int":          &MapType{Key: StringType, Value: IntType},
		"chan int":                &ChanType{Dir: SEND | RECV, Elt: IntType},
		"<-chan int":              &ChanType{Dir: SEND, Elt: IntType},
		"chan<- int":              &ChanType{Dir: RECV, Elt: IntType},
		"struct{A int; b string}": &StructType{Fields: []FieldType{{Name: "A", Type: IntType}, {Name: "b", Type: StringType}}},
		"interface {M func()}":    &InterfaceType{Methods: []FieldType{{Name: "M", Type: &FuncType{}}}},
		"<X>{}":                   &InterfaceType{Generic: "X"},
		"func()":                  &FuncType{},
		"func(int, string) int":   &FuncType{Params: []FieldType{{Type: IntType}, {Type: StringType}}, Results: []FieldType{{Type: IntType}}},
		"func(int) (int, string)": &FuncType{Params: []FieldType{{Type: IntType}}, Results: []FieldType{{Type: IntType}, {Type: StringType}}},
		"(int,string)":            &tupleType{Elts: []Type{IntType, StringType}},
		"F int \"t\"":             FieldType{Name: "F", Type: IntType, Tag: "t"},
		"(embedded) int":          FieldType{Type: IntType},
	} {
		if got := typ.String(); got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	}
}

func TestTypeTextClipsAtTheCap(t *testing.T) {
	sized := func(nameLen int) Type {
		return &StructType{Fields: []FieldType{{Name: Name(strings.Repeat("F", nameLen)), Type: IntType}}}
	}
	// struct{<name> int} adds 12 bytes to the name.
	exact := sized(maxTypeTextBytes - 12).String()
	if len(exact) != maxTypeTextBytes {
		t.Errorf("text at the cap: len %d", len(exact))
	}
	over := sized(maxTypeTextBytes - 11).String()
	if len(over) != maxTypeTextBytes+len(typeTextClipMark) {
		t.Errorf("text past the cap: len %d", len(over))
	}
}

// The cut lands on a rune boundary wherever a multi-byte name crosses the cap.
func TestTypeTextCutKeepsUTF8(t *testing.T) {
	for pad := range 3 {
		name := strings.Repeat("F", pad) + strings.Repeat("\u0800", maxExprTypeTextBytes/3+1)
		text := BoundedTypeText(&StructType{Fields: []FieldType{{Name: Name(name), Type: IntType}}}, maxExprTypeTextBytes)
		if !utf8.ValidString(text) {
			t.Errorf("pad %d: invalid UTF-8 at the cut", pad)
		}
		if got, want := len(text), maxExprTypeTextBytes+len(typeTextClipMark); got > want || got < want-2 {
			t.Errorf("pad %d: len %d, want at most 2 under %d", pad, got, want)
		}
	}
}

// Once a cut backs off to a rune boundary, the frames around it write nothing into
// the bytes left.
func TestTypeTextCutWritesNothingAfter(t *testing.T) {
	const r = "\u0800"
	for pad := range 3 {
		name := strings.Repeat("F", pad) + strings.Repeat(r, maxExprTypeTextBytes/3+1)
		fields := []FieldType{{Name: Name(name), Type: IntType}, {Name: "G", Type: IntType}}
		text := BoundedTypeText(&StructType{Fields: fields}, maxExprTypeTextBytes)
		if !strings.HasSuffix(text, r+typeTextClipMark) {
			t.Errorf("pad %d: text ends %q", pad, text[len(text)-8:])
		}
	}
}

// 20000 int parameters print as about 100 KB of text; the type is built in Go, so
// no sealer refuses it first.
func TestWriteProtectedCutsTheTypeText(t *testing.T) {
	params := make(FieldTypeList, 20000)
	for i := range params {
		params[i] = FieldType{Type: IntType}
	}
	tv := TypedValue{T: &FuncType{Params: params}}

	var buf bytes.Buffer
	mw := newUnmeteredWriter(&buf)
	tv.WriteProtected(mw, newSeenValues())
	mw.Flush()

	if n := buf.Len(); n > maxTypeTextBytes+64 {
		t.Fatalf("printed %d bytes, want at most about %d", n, maxTypeTextBytes)
	}
}

// A type read back from the store keeps an id past the cap, in every kind fillType
// sees, and the same type built new is refused.
func TestStoredTypeKeepsItsID(t *testing.T) {
	structOfID := func(n int) *StructType {
		return &StructType{Fields: []FieldType{{Name: Name(strings.Repeat("A", n-12)), Type: IntType}}}
	}
	atCap := func() Type { return structOfID(maxTypeIDBytes) }
	half := func() Type { return structOfID(maxTypeIDBytes/2 + 8) }
	field := func(typ Type) FieldType { return FieldType{Name: "F", Type: typ} }
	for name, build := range map[string]func() Type{
		"pointer":                func() Type { return &PointerType{Elt: atCap()} },
		"array":                  func() Type { return &ArrayType{Len: 2, Elt: atCap()} },
		"slice":                  func() Type { return &SliceType{Elt: atCap()} },
		"map":                    func() Type { return &MapType{Key: atCap(), Value: IntType} },
		"struct":                 func() Type { return &StructType{Fields: []FieldType{field(atCap())}} },
		"struct, in its list":    func() Type { return &StructType{Fields: []FieldType{field(half()), field(half())}} },
		"func":                   func() Type { return &FuncType{Params: []FieldType{field(atCap())}} },
		"params, in their list":  func() Type { return &FuncType{Params: []FieldType{field(half()), field(half())}} },
		"results, in their list": func() Type { return &FuncType{Results: []FieldType{field(half()), field(half())}} },
		"tuple":                  func() Type { return &tupleType{Elts: []Type{atCap()}} },
		"interface": func() Type {
			return &InterfaceType{Methods: []FieldType{field(&FuncType{Params: []FieldType{field(atCap())}})}}
		},
		"methods, in their list": func() Type {
			method := func(name Name) FieldType {
				return FieldType{Name: name, Type: &FuncType{Params: []FieldType{field(half())}}}
			}
			return &InterfaceType{Methods: []FieldType{method("M"), method("N")}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			func() {
				defer func() {
					if r := recover(); !strings.Contains(fmt.Sprint(r), "type too large") {
						t.Errorf("built new: recovered %v", r)
					}
				}()
				build().TypeID()
			}()
			if id := fillType(nil, build()).TypeID(); len(id) <= maxTypeIDBytes {
				t.Errorf("read back: id of %d bytes", len(id))
			}
		})
	}
}

// Loading a map computes each key's id, and saving an object checks its types, so
// a stored key past the cap must pass both.
func TestStoredMapKeyPastTheCap(t *testing.T) {
	ds := NewStore(NewAllocator(1<<30), nil, nil)
	keyType := &PointerType{Elt: &StructType{Fields: []FieldType{{Name: Name(strings.Repeat("A", maxTypeIDBytes)), Type: IntType}}}}
	mv := &MapValue{List: &MapList{}}
	mv.List.Append(nil, TypedValue{T: keyType}).Value = typedString("v")
	fillTypesOfValue(nil, ds, mv)
	if len(mv.vmap) != 1 {
		t.Errorf("vmap holds %d keys", len(mv.vmap))
	}
	rlm := &Realm{Path: "gno.land/r/test/stored"}
	rlm.assertTypeIsPublic(ds, &FuncType{Params: []FieldType{{Name: "p", Type: keyType}}, stored: true}, map[TypeID]struct{}{})
}

// Text runs past its id by at most one byte per list item, so a type under the
// id's cap never loses its text to the text's cap.
func TestTypeTextWholeUnderTheIDCap(t *testing.T) {
	var params []FieldType
	for range (maxTypeIDBytes - 8) / 4 {
		params = append(params, FieldType{Type: IntType})
	}
	ft := &FuncType{Params: params}
	if id := ft.TypeID(); len(id) < maxTypeIDBytes-8 {
		t.Fatalf("id: len %d", len(id))
	}
	if text := ft.String(); strings.HasSuffix(text, typeTextClipMark) {
		t.Errorf("text of %d bytes was cut", len(text))
	}
}

// Once the text is full the walk leaves the rest of the type unvisited, so a
// type of 8^8 leaves allocates what one cap costs, and the nil type of its last
// field, far past the cut, is never reached: a walk that ran on would panic on it.
func TestTypeTextStopsAtTheCap(t *testing.T) {
	var typ Type = IntType
	for range 8 {
		fields := make([]FieldType, 8)
		for i := range fields {
			fields[i] = FieldType{Name: "F", Type: &ArrayType{Len: 1 << 40, Elt: typ}}
		}
		typ = &StructType{Fields: fields}
	}
	st := typ.(*StructType)
	st.Fields = append(st.Fields, FieldType{Name: "Z"})
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_ = typ.String()
	runtime.ReadMemStats(&after)
	if n := after.TotalAlloc - before.TotalAlloc; n >= 8*maxTypeTextBytes {
		t.Errorf("text allocated %d bytes", n)
	}
}

// Four source bytes can repeat a type inside an expression's text.
func TestExprTextCutsAType(t *testing.T) {
	sized := func(n int) Type { return typeIDOfLen(n) }
	typeval := func(n int) string {
		return ConstExpr{TypedValue: TypedValue{T: gTypeType, V: TypeValue{Type: sized(n)}}}.String()
	}
	if got := len(typeval(maxExprTypeTextBytes)); got != maxExprTypeTextBytes+len("typeval{}") {
		t.Errorf("typeval at the cap: len %d", got)
	}
	if got := len(typeval(maxExprTypeTextBytes + 1)); got != maxExprTypeTextBytes+len("typeval{}")+len(typeTextClipMark) {
		t.Errorf("typeval past the cap: len %d", got)
	}
	constant := func(n int) string {
		return ConstExpr{TypedValue: TypedValue{T: sized(n)}}.String()
	}
	if got := len(constant(maxExprTypeTextBytes)); got != maxExprTypeTextBytes+len("(const (0 ))") {
		t.Errorf("constant at the cap: len %d", got)
	}
	if got := len(constant(maxExprTypeTextBytes + 1)); got != maxExprTypeTextBytes+len("(const (0 ))")+len(typeTextClipMark) {
		t.Errorf("constant past the cap: len %d", got)
	}
	// A value's text outside an expression keeps its declared type whole.
	if got := len(TypedValue{T: sized(maxExprTypeTextBytes + 1)}.String()); got != maxExprTypeTextBytes+1+len("(0 )") {
		t.Errorf("value's text: len %d", got)
	}
	// A type value is cut on every route, outside an expression too.
	if got := len(TypedValue{T: gTypeType, V: TypeValue{Type: sized(maxExprTypeTextBytes + 1)}}.String()); got != maxExprTypeTextBytes+len("(typeval{} type{})")+len(typeTextClipMark) {
		t.Errorf("type value's text: len %d", got)
	}
	consttype := func(n int) string { return constTypeExpr{Type: sized(n)}.String() }
	if got := len(consttype(maxExprTypeTextBytes)); got != maxExprTypeTextBytes+len("(const-type )") {
		t.Errorf("const-type at the cap: len %d", got)
	}
	if got := len(consttype(maxExprTypeTextBytes + 1)); got != maxExprTypeTextBytes+len("(const-type )")+len(typeTextClipMark) {
		t.Errorf("const-type past the cap: len %d", got)
	}
}

// vm/qeval skips type-checking, so it can build `struct{ A }` over a named
// type that points back to itself; building it must return, not recurse.
func TestReachableSurface_PointerCycle(t *testing.T) {
	for _, decls := range []string{"type A *A", "type A *B\ntype B *A"} {
		for _, expr := range []string{"struct{ A }{}", "struct{ *A }{}"} {
			m := NewMachine("testdata", nil)
			m.RunFiles(m.MustParseFile("testdata.gno", "package testdata\n"+decls+"\n"))
			if res := m.Eval(m.MustParseExpr(expr)); len(res) != 1 {
				t.Fatalf("%q over %q: got %d results, want 1", expr, decls, len(res))
			}
			m.Release()
		}
	}
}

// A named type reached through two embedded fields counts once per field, so
// the cut fires only on a cycle. qeval accepts the duplicate field name T. The
// []int base reaches the walk's default case, which keeps T's methods too.
func TestReachableSurface_RepeatIsNotCycle(t *testing.T) {
	n := MaxInterfaceMethods/2 + 1 // over the cap only when counted twice
	for _, base := range []string{"struct{}", "[]int"} {
		var src strings.Builder
		src.WriteString("package testdata\ntype T " + base + "\n")
		for i := range n {
			fmt.Fprintf(&src, "func (T) M%d() {}\n", i)
		}
		m := NewMachine("testdata", nil)
		m.RunFiles(m.MustParseFile("testdata.gno", src.String()))
		func() {
			defer m.Release()
			defer func() {
				want := fmt.Sprintf("has %d promoted methods", 2*n)
				if r := recover(); r == nil || !strings.Contains(fmt.Sprint(r), want) {
					t.Fatalf("type T %s: want a panic naming %q, got %v", base, want, r)
				}
			}()
			m.Eval(m.MustParseExpr("struct{ T; *T }{}"))
		}()
	}
}

// A type declared in a nested block is keyed to the block's own location, so a
// block without one is refused: a block the preprocessor built has a span
// derived from its parent's, and a block evaluated outside a file has none.
func TestDeclareWith_NestedBlockNeedsOwnLocation(t *testing.T) {
	const pkgPath = "gno.land/r/test/blocks"
	loc := Location{PkgPath: pkgPath, File: "a.gno", Span: Span{Pos: Pos{Line: 3, Column: 2}, End: Pos{Line: 6, Column: 3}}}
	parsed := &BlockStmt{}
	parsed.SetLocation(loc)
	if got := declareWith(pkgPath, parsed, "t", IntType).ParentLoc; got != loc {
		t.Fatalf("parsed block: ParentLoc = %v, want %v", got, loc)
	}

	derived := loc
	derived.Num = 1
	for _, bad := range []Location{{}, derived} {
		bs := &BlockStmt{}
		bs.SetLocation(bad)
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("block at %q: declareWith did not panic", bad)
				}
			}()
			declareWith(pkgPath, bs, "t", IntType)
		}()
	}

	m := NewMachine("testdata", nil)
	defer m.Release()
	m.RunFiles(m.MustParseFile("testdata.gno", "package testdata\n"))
	defer func() {
		want := "expected type expr but got *gnolang.BlockStmt"
		if r := recover(); !strings.Contains(fmt.Sprint(r), want) {
			t.Fatalf("Eval: want a panic naming %q, got %v", want, r)
		}
	}()
	m.Eval(m.MustParseExpr("func() int { { type t int; return int(t(1)) } }()"))
}
