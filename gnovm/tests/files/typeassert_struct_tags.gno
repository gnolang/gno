package main

// NOTE: differs from Go, where the first assertion fails: runtime value type
// identity in the GnoVM is TypeID-based, so struct tags are not part of a
// value's dynamic type. Tags are enforced statically only.
// See gnovm/adr/pr5785_type_identity.md.

func main() {
	var i interface{} = struct {
		A int `a`
	}{1}
	_, ok := i.(struct {
		A int `b`
	})
	println(ok)
	_, ok = i.(struct {
		A int `a`
	})
	println(ok)
}

// Output:
// true
// true
