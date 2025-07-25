package main

func main() {
	for i, v := range nil {
		println(i, v)
	}
}

// Error:
// main/range8.gno:4:2-6:3: cannot range over nil

// TypeCheckError:
// main/range8.gno:4:20: cannot range over nil
