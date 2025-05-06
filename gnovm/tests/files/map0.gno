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

// TypeCheckError:
// main/files/map0.gno:7:3: duplicate key "hello" in map literal; main/files/map0.gno:4:2: declared and not used: m
