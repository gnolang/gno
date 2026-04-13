package main

type A interface {
    F()
}

type B interface {
    F()
    G()
}

type T struct{}

func (T) F() {}

func (T) G() {}

func main() {
    var a A = T{}
    var b B = T{}

    println(b == a)
    println(a == b)
}

// Output:
// true
// true
