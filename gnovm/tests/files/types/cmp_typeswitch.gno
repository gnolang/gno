package main

func main() {
	var l interface{}
	l = int64(0)

	switch val := l.(type) {
	case int64, int:
		if val == 0 {
			println("l is zero")
		} else {
			println("NOT zero")
		}
	}
}

// Output:
// NOT zero
