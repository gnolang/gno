package crossrealm

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
