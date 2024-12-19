package main

func main() {
	for i, v := range nil {
		println(i, v)
	}
}

// Error:
// main/files/range8.gno:4:2: cannot range over nil
