package main

type Runner interface {
	Run()
}

type Swimmer interface {
	Swim()
}

func main() {
	a := Runner(nil)
	println(Swimmer(a))
}

// Error:
// main/files/interface47.gno:13:10: main.Runner does not implement main.Swimmer (missing method Swim)
