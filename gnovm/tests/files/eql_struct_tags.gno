package main

// NOTE: differs from Go, which prints false: runtime value type identity in
// the GnoVM is TypeID-based, so struct tags are not part of a value's
// dynamic type. Tags are enforced statically only.
// See gnovm/adr/pr5785_type_identity.md.

func main() {
	var i interface{} = struct {
		A int `a`
	}{1}
	var j interface{} = struct {
		A int `b`
	}{1}
	println(i == j)
}

// Output:
// true
