package main

func main() {
	m := map[string]string{
		"hello": "foo",
		"world": "bar",
		"hello": "",
	}
}

// TypeCheckError:
// main/map0.gno:7:3: duplicate key "hello" in map literal; main/map0.gno:4:2: declared and not used: m
