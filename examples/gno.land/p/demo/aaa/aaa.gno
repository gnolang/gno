// Experiment with eliding

package aaa

type Foo struct {
	a string
	b int
}

type Bar struct {
	c string
	d Foo
}

func Experiment() {
	// a := [3]T{0: "a", 1: "b", 2: "c"}
	// var a []Foo = []Foo{{"A", 1}, {"B", 2}}
	var b []Bar = []Bar{
		{"A", {"A", 1}},
		{"B", {"B", 2}},
	}
	// _ = a
	_ = b
}
