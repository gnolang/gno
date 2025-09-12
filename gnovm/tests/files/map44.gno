// PKGPATH: gno.land/r/test
package test

import "fmt"

type Foo struct {
	name string
}

var m map[string]Foo

func init() {
	m = make(map[string]Foo)
	m["hey"] = Foo{name: "yep"}
}

func main() {
	m["hey"] = Foo{name: "what?"}
	fmt.Println(m)
}

// Output:
// map[hey:{what?}]
