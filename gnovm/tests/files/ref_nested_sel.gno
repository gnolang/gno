package main

// Taking the address of a deeply-nested selector (&a.b.c) should heap-escape
// the root variable `a`, not just the intermediate `a.b`.

type Inner struct {
	val int
}

type Outer struct {
	inner Inner
}

var gp *int

func f() {
	o := Outer{inner: Inner{val: 42}}
	gp = &o.inner.val
}

func main() {
	f()
	println(*gp)
	*gp = 99
	println(*gp)
}

// Preprocessed:
// file{ package main; type Inner (const-type main.Inner); type Outer (const-type main.Outer); var gp<!~VPBlock(2,2)> *((const-type int)); func f() { o<!~VPBlock(1,0)> := (const-type main.Outer){inner<VPField(0,0,inner)>: (const-type main.Inner){val<VPField(0,0,val)>: (const (42 int))}}; gp<~VPBlock(3,2)> = &(o<~VPBlock(1,0)>.inner.val) }; func main() { f<VPBlock(3,3)>(); (const (println func(...interface {})))(*(gp<~VPBlock(3,2)>)); *(gp<~VPBlock(3,2)>) = (const (99 int)); (const (println func(...interface {})))(*(gp<~VPBlock(3,2)>)) } }

// Output:
// 42
// 99
