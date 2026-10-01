// A plain (unqualified) embedded type alias keeps the alias name as the field
// name: struct{Int} stays distinct from struct{int}, and two same-named aliases
// of different underlying types stay distinct too. (alias3 covers the qualified
// selector form pkg.T, plus pointer and interface embeds.)
package main

type Int = int

type A = struct{ int }
type B = struct{ Int }

func main() {
	var x, y interface{} = A{}, B{}
	if x == y {
		mustNot()
	}

	{
		type C = int32
		x = struct{ C }{}
	}
	{
		type C = uint32
		y = struct{ C }{}
	}
	if x == y {
		mustNot()
	}
}
func mustNot() { panic(badMsg) }

const badMsg = "FAIL"
