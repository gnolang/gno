package main

import "fmt"

type Foo struct{ Name string }

func (f *Foo) Bar(param string) string {
	fmt.Printf("%+v %+v\n", f, param)
	return "ok"
}

func main() {
	const a = (*Foo).Bar(&Foo{Name: "value"}, "param")
}
