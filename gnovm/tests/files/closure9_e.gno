package main

func main() {
	for i := 0; i < 5; i++ {
		if i == 1 {
			panic("error 1")
		}
	}
}

// Error:
// error 1
