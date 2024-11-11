package main

import "fmt"

func main() {
	counter0 := 0
	counter1 := 0

	y := 0

	var fs []func()

	defer func() {
		for _, ff := range fs {
			ff()
		}
	}()

LOOP_START:
	if counter0 < 2 {
		x := y
		counter1 = 0
		fmt.Printf("Outer loop start: counter0=%d\n", counter0)

	NESTED_LOOP_START:
		if counter1 < 2 {
			fmt.Printf("  Nested loop: counter1=%d\n", counter1)
			fs = append(fs, func() { println(x) })

			counter1++
			goto NESTED_LOOP_START
		}

		fmt.Println("Exiting nested loop")
		counter0++
		y++
		goto LOOP_START
	} else {
		return
	}
}

// Preprocessed:
// file{ package main; import fmt fmt; func main() { counter0<!VPBlock(1,0)> := (const (0 int)); counter1<!VPBlock(1,1)> := (const (0 int)); y<!VPBlock(1,2)> := (const (0 int)); var fs<!VPBlock(1,3)> []func(); defer func func(){ for _<VPBlock(0,0)>, ff<VPBlock(1,0)> := range fs<VPBlock(3,3)> { ff<VPBlock(1,0)>() } }(); if counter0<VPBlock(2,0)> < (const (2 int)) { x<!~VPBlock(1,0)> := y<VPBlock(2,2)>; counter1<VPBlock(2,1)> = (const (0 int)); fmt<VPBlock(3,0)>.Printf((const ("Outer loop start: counter0=%d\n" string)), (const-type gonative{interface {}})(counter0<VPBlock(2,0)>)); if counter1<VPBlock(3,1)> < (const (2 int)) { fmt<VPBlock(4,0)>.Printf((const ("  Nested loop: counter1=%d\n" string)), (const-type gonative{interface {}})(counter1<VPBlock(3,1)>)); fs<VPBlock(3,3)> = (const (append func(x []func()(),args ...func()())(res []func()())))(fs<VPBlock(3,3)>, func func(){ (const (println func(xs ...interface{})()))(x<~VPBlock(1,0)>) }<x<()~VPBlock(2,0)>>); counter1<VPBlock(3,1)>++; goto NESTED_LOOP_START<1,3> }; fmt<VPBlock(3,0)>.Println((const ("Exiting nested loop" string))); counter0<VPBlock(2,0)>++; y<VPBlock(2,2)>++; goto LOOP_START<1,5> } else { return } } }

// Output:
// Outer loop start: counter0=0
//   Nested loop: counter1=0
//   Nested loop: counter1=1
// Exiting nested loop
// Outer loop start: counter0=1
//   Nested loop: counter1=0
//   Nested loop: counter1=1
// Exiting nested loop
// 0
// 0
// 1
// 1
