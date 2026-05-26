package main

func shouldPanic(f func()) {
	defer func() {
		r := recover()
		if r == nil {
			panic("not panicking")
		}
		println("recovered:", r)
	}()
	f()
}

func f() {
	length := int(^uint(0) >> 1)
	a := make([]struct{}, length)
	b := make([]struct{}, length)
	_ = append(a, b...)
}

func main() {
	shouldPanic(f)
	println("ok")
}

// Output:
// recovered: runtime error: makeslice: len out of range
// ok
