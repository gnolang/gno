package main

func main() {
	const s = ""
	t := map[any]int{len(s) + 0: 0, 0: 1}
	_ = t // keep the variable used; the test is about something else
}

// TypeCheckError:
// main/map42.gno:5:34: duplicate key 0 in map literal
