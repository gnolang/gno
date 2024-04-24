package main

func main() {
	func() {
		var i string = "foobar"
		println(i)
	}()
}

// Output:
// foobar
