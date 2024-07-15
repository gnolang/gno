package main

func main() {

	for i := 0; i < 3; i++ {

		defer func() { println(i) }()

	}

}

// Output:
// 3
// 3
// 3
