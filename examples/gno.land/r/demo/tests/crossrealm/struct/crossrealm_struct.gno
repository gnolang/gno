package crossrealm_struct

type Bar struct {
	A int
}

var bar *Bar

// TODO: deprovision this
var Bar2 *Bar = &Bar{A: 22} // exported

func (b *Bar) String() {
	println("b.A: ", b.A)
}

func SetBar(b *Bar) *Bar {
	bar = b
	return bar
}

func ChangeBar() {
	Bar2.A += 1
	//println(Bar2.A)
}

func (b *Bar) ModifyBar() {
	b.A += 1
}

// -------------------------------------------------
type A struct {
	name string
}

type B struct {
	name string
}

type D struct {
	name string
}

type C struct {
	A
	B
	D
}

var a A = A{name: "a"}
var b B = B{name: "b"}

var c *C

func init() {
	c = &C{}
	c.A = a
	c.B = b
}

func GetStruct(cb func(v *C)) *C {
	c.D = D{name: "d"} // this is not attached before return
	cb(c)
	return c
}
