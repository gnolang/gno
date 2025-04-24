package main

type A int

func main() {
	println(A(1))
	type A string
	println(A("a"))
}

// Output:
// (1 main.A)
// ("a" main[main/files/type42.gno:5:1].A)

// Preprocessed:
// file{ package main; type A (const-type main.A); func main() { (const (println func(...interface {})))((const (1 main.A))); type A (const-type main[main/files/type42.gno:5:1].A); (const (println func(...interface {})))((const ("a" main[main/files/type42.gno:5:1].A))) } }
