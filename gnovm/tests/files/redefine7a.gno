package main

func main() {
	for i := 0; i < 3; i++ {
		func() {
            i += 1
        }()
		println(i)
	}
}

// Output:
// 0
// 1
// 2
