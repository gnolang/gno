package main

func main() {
	type s struct{}
	m := map[s]int{
	_ = m // keep the variable used; the test is about something else
		{}: 1,
		{}: 2,
	}
	// NOTE(tb): surprisingly, this code doesn't trigger any error in go, whereas
	// the resulting map has only one item (the last one).
}

// TypeCheckError:
// main/map1.gno:6:4: expected '==', found '='

// Error:
// map1.gno:6:4: expected '==', found '='
