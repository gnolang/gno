package main

func main() {
	const s = ""
	t := map[any]int{len(s) + 0: 0, 0: 1}
}

// TypeCheckError:
// main/map42.gno:5:34: duplicate key 0 in map literal; main/map42.gno:5:2: declared and not used: t
