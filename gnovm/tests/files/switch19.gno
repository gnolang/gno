package main

import "fmt"

type fii interface {
	Hello()
}

type Bir struct{}

func (b Bir) Yo() {
	fmt.Println("Yo", b)
}

func (b Bir) Hello() {
	fmt.Println("Hello", b)
}

type Boo struct {
	Name string
}

func (b Boo) Hello() {
	fmt.Println("Hello", b)
	fmt.Println(b.Name)
}

type Bar struct{}

func (b Bar) Hello() { fmt.Println("b:", b) }

func inCall(foo fii) {
	fmt.Println("inCall")
	switch a := foo.(type) {
	case Boo, Bir:
		a.Hello()
	case Bir:
		a.Yo()
	default:
		fmt.Println("a:", a)
	}
}

func main() {
	boo := Bir{}
	inCall(boo)
	inCall(Bar{})
}

// Error:
// main/switch19.gno:34:2-41:3: duplicate type main.Bir in type switch

// TypeCheckError:
// main/switch19.gno:37:7: duplicate case Bir in type switch; main/switch19.gno:35:12: 	previous case
