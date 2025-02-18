// PKGPATH: gno.land/r/crossrealm_map
package crossrealm_map

type Foo struct{}

var f = Foo{}

var m map[string]int

func init() {
	m = make(map[string]int)
	m["a"] = 1
}

func GetMap() map[string]int {
	return m
}

func GetMap2(f func(map[string]int)) {
	m1 := map[string]int{"a": 1}
	f(m1)
}

func GetMap4() map[string]Foo {
	var m1 = map[string]Foo{"foo": f}
	return m1
}

// --------------------------------------
type MyMap map[string]int

var m2 = MyMap{"a": 1}

func GetMap3() MyMap {
	return m2
}

// TODO: add some map with elements of object

// Realm:
