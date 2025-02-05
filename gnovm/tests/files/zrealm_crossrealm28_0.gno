// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

type foo struct {
	name string
}

var root interface{}

func main() {
	var arr []foo // unreal array with type from local realm
	arr = append(arr, foo{name: "1"})
	arr = append(arr, foo{name: "2"})
	arr = append(arr, foo{name: "3"})

	fs := arr[1:]
	root = fs
	println("ok")
}

// Output:
// ok
