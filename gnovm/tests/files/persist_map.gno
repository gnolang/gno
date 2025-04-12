// PKGPATH: gno.land/r/demo/tests_test
package tests_test

var amap map[string]string = map[string]string{"a": "1"}

func init() {
	println("preinit", amap)
	amap["b"] = "2"
	println("postinit", amap)
}

func main() {
	crossing()

	println("premain", amap)
	amap["c"] = "3"
	println("postmain", amap)
}

// Output:
// preinit map{("a" string):("1" string)}
// postinit map{("a" string):("1" string),("b" string):("2" string)}
// premain map{("a" string):("1" string),("b" string):("2" string)}
// postmain map{("a" string):("1" string),("b" string):("2" string),("c" string):("3" string)}
