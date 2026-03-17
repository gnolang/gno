// PKGPATH: gno.land/r/test
package test

type Bar struct{
	name string
}

type Foo struct{
	bar Bar
	num int
}

var f = Foo{bar: Bar{name: "b"}, num: 1}

func main(){
	f.num = 2 // make f dirty.
	f.bar.name = "c" // make b dirty, thus f has a dirty child. dirty children should be allowed in `toRefValue()`
	println(f)
}

// Output:
// (struct{(struct{("c" string)} gno.land/r/test.Bar),(2 int)} gno.land/r/test.Foo)