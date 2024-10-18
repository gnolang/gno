package crossrealm

type Bar struct {
	A int
}

var bar *Bar
var Bar2 *Bar // exported

func SetBar(b *Bar) *Bar {
	bar = b
	return bar
}

func CallBar() {
	bar.A += 1
	println(bar.A)
}
