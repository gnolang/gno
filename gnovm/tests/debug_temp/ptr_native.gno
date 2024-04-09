package main

import "time"

type MyTime *time.Time

type Foo struct {
	age  int
	name string
	time MyTime
}

func main() {
	f := Foo{age: 1, name: "hello", time: &time.Time{}}
	println(f.time)
}

// Output:
