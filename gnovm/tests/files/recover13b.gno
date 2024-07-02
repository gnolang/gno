package main

type s struct {
	el func()
}

func p() {
	var el *s
	defer func() {
		if r := recover(); r != nil {
			println(r)
		}
	}()
	el.el()
}

func main() {
	p()
}

// Output:
// interface conversion: gnolang.Value is nil, not gnolang.PointerValue
