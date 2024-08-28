package main

type zilch interface{}

func main() {
	_ = zilch(nil)
}

// Output:
