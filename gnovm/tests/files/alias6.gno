// A struct embedding an interface alias takes the field name from the written
// spelling (the struct rule), NOT the resolved interface — so struct{ SAlias }
// (field "SAlias") is distinct from struct{ Stringer } (field "Stringer"),
// even though SAlias = Stringer. This is the opposite of the interface-embed
// rule (alias5/iface_embed_*), confirming the spelling-vs-flatten split is
// dispatched by container kind, not by the embedded type's kind.
package main

type Stringer interface{ Str() string }
type SAlias = Stringer

type T struct{}

func (T) Str() string { return "hi" }

func main() {
	var x, y interface{} = struct{ SAlias }{}, struct{ Stringer }{}
	if x == y {
		panic("FAIL: struct{SAlias} == struct{Stringer}")
	}
	var z interface{} = struct{ SAlias }{}
	if x != z {
		panic("FAIL: struct{SAlias} != struct{SAlias}")
	}
	// the embedded field is named SAlias and promotes the method set
	s := struct{ SAlias }{SAlias: T{}}
	println(s.Str())
}

// Output:
// hi
