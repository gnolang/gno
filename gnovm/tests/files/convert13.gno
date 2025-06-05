package main

func main() {
	b := struct{ name string }([]byte("hello"))
}

// Error:

// TypeCheckError:
