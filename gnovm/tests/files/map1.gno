package main

func main() {
	type s struct{}
	m := map[s]int{
		{}: 1,
		{}: 2,
	}
	_ = m // keep the variable used; the test is about something else
	// NOTE(tb): surprisingly, this code doesn't trigger any error in go, whereas
	// the resulting map has only one item (the last one).
}
