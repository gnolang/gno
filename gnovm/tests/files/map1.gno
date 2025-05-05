package main

func main() {
	type s struct{}
	m := map[s]int{
		{}: 1,
		{}: 2,
	}
	// NOTE(tb): surprisingly, this code doesn't trigger any error in go, whereas
	// the resulting map has only one item (the last one).
}

// TypeCheckError:
// main/files/map1.gno:5:2: declared and not used: m
