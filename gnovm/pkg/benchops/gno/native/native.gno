package native

func Print_1() {
	b := ""
	for i := 0; i < 1; i++ {
		b += "a"
	}
	print(b)
}

func Print_1000() {
	b := ""
	for i := 0; i < 1000; i++ {
		b += "a"
	}
	print(b)
}

func Print_10000() {
	b := ""
	for i := 0; i < 10000; i++ {
		b += "a"
	}
	print(b)
}
