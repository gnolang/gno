package main

import (
	"fmt"
)

type Cheese struct {
	Property string
}

func (t *Cheese) Hello(param string) {
	fmt.Printf("%+v %+v", t, param)
}

func main() {
	(*Cheese).Hello(&Cheese{Property: "value"}, "param")
}

// Output:
// &{Property:value} param
