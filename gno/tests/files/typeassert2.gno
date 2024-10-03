package main

type ex int

func (ex) Error() string { return "" }

type A interface {
	Do(s string)
}

func main() {
	defer func() {
		e := recover()
		if _, ok := e.(ex); ok {
			println("wat")
		} else {
			println("ok")
		}
	}()
	defer func() {
		e := recover()
		if _, ok := e.(A); ok {
			println("wat")
		} else {
			println("ok")
		}
	}()
}

// Output:
// ok
// ok
