package pkg

func Hello() {
	println("called Hello")
}

func World() {
	println("called World")
}

func hello() {
	println("called hello")
}

func WithArg(i int) {
	switch i {
	case 0:
		println("zero")
	case 1:
		println("one")
	case 2:
		println("two")
	default:
		println("out of range! :S")
	}
}
