// An embedded interface is named from its resolved type, not the written
// spelling, so embedding an alias is identical to embedding its target:
// interface{ SAlias } == interface{ Stringer }. This is the inverse of the
// struct rule (alias2-4), because Go computes interface identity from the
// method set, not the embedded name. Guards against the regression where
// alias-named embeds got distinct TypeIDs.
//
// NOTE: this only equalizes alias-vs-target. Structurally identical
// interfaces with different spellings still diverge (interface{ Stringer }
// != interface{ Str() string }); fixing that needs method-set flattening
// (follow-up, branch scratch/iface-flatten).
package main

type Stringer interface{ Str() string }
type SAlias = Stringer

func main() {
	var x, y interface{} = struct{ A interface{ SAlias } }{}, struct{ A interface{ Stringer } }{}
	println("alias==canonical:", x == y)
}

// Output:
// alias==canonical: true
