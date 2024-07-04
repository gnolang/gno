package main

type Count struct {
	Num int
}

func main() {
	c := &Count{
		1,
	}
	(*c).Num += 1
}
