package main

func main() {
	{
		b := []byte("ABCDEFGHIJKL")
		a := b
		println(&b[0] == &a[0], b[0], a[0])

		// modifying the underlying array modifies both a[0] and b[0],
		// as it should
		a[0] = 11
		println(a[0], b[0])
	}

	{
		b := []byte{1, 2, 3}
		a := b
		println(&b[0] == &a[0], b[0], a[0])

		// modifying the underlying array modifies both a[0] and b[0],
		// as it should
		a[0] = 11
		println(a[0], b[0])
	}
}

// Output:
// true 65 65
// 11 11
// true 1 1
// 11 11
