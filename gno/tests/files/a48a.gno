package main

func main() {
	{
		b := []byte("ABCDEFGHIJKL")
		a := b
		println(&a[0] == &a[0])
	}
}

// Output:
// true
