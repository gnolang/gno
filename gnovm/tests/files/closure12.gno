package main

func trymake() func() {
	switch i := 1; i {
	case 1:
		return func() {
			println(i)
		}
	default:
	}
	return nil
}
func main() {
	trymake()()
}

// Output:
// 1
