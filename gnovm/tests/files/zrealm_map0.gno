// PKGPATH: gno.land/r/test
package test

var m map[string]int

func init() {
	m = make(map[string]int)
}

func main(cur realm,) {
	m["foobar"] = 1
	println(m)
}

// Output:
// map{("foobar" string):(1 int)}
