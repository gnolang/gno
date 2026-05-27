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
