// An embedded interface contributes only its method set to identity, never
// the written spelling, so embedding an alias is identical to embedding its
// target: interface{ SAlias } == interface{ Stringer }. This is the inverse
// of the struct rule (alias2-4), because Go computes interface identity from
// the method set, not the embedded name. See iface_embed_id.gno for the wider
// flattening cases (embed-vs-explicit, diamond).
package main

type Stringer interface{ Str() string }
type SAlias = Stringer

func main() {
	var x, y interface{} = struct{ A interface{ SAlias } }{}, struct{ A interface{ Stringer } }{}
	println("alias==canonical:", x == y)
}

// Output:
// alias==canonical: true
