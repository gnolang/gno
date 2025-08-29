package main

type Key struct {
	num int
}

func main() {
	k := Key{num: 1}
	m := map[Key]int{k: 100}
	println(m)

	k.num = 2

	println(m)
	println(k)
}

// Output:
// map{(struct{(1 int)} main.Key):(100 int)}
// map{(struct{(1 int)} main.Key):(100 int)}
// (struct{(2 int)} main.Key)
