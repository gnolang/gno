package main

func main() {
	m := map[string]string{
		"hello": "foo",
		"world": "bar",
		"hello": "",
	}
}

// Error:
// duplicate key "hello" in map literal
