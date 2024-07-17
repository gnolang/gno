package main

func main() {
	i := 0

	for true {
		println(i)
		i++
		if i == 5 {
			break
		}
	}
}

// Output:
// 0
// 1
// 2
// 3
// 4
