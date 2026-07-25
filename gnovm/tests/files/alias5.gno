// Uverse builtin aliases follow the same embedded-field naming rule as user
// aliases (see alias2): the field is named by the alias as written, so
// struct{rune} stays distinct from struct{int32}, and struct{byte} from
// struct{uint8}. Matches Go.
package main

func main() {
	var x, y interface{} = struct{ rune }{}, struct{ int32 }{}
	if x == y {
		mustNot()
	}
	x, y = struct{ byte }{}, struct{ uint8 }{}
	if x == y {
		mustNot()
	}
}
func mustNot() { panic(badMsg) }

const badMsg = "FAIL"
