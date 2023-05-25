// PKGPATH: gno.land/r/demo/tests_test
package tests_test

var amap         map[string]string = map[string]string{"a": "1"}
	
func init() {
	println("preinit", amap)
	amap["b"] = "2"
	println("postinit", amap)
}

func main() {
	println("premain", amap)
	amap["b"] = "2"
	println("postmain", amap)
}



// Current behavior: panics
