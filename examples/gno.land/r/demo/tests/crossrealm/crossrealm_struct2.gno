package crossrealm

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
