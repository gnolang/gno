package main

// NOTE: differs from Go, which prints "default": runtime value type identity
// in the GnoVM is TypeID-based, so struct tags are not part of a value's
// dynamic type. Tags are enforced statically only.
// See gnovm/adr/pr5785_type_identity.md.

func main() {
	var i interface{} = struct {
		A int `a`
	}{1}
	switch i.(type) {
	case struct {
		A int `b`
	}:
		println("b")
	default:
		println("default")
	}
}

// Output:
// b
