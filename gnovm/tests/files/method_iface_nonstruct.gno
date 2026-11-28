// Non-struct declared receivers (int, array) through a lazy interface bind:
// these shapes cannot embed, so resolution records no cycle identity and
// terminates immediately — also as the operand of a nested hop. Matches Go.
package main

type IG interface{ Get() int }

type MyInt int

func (m MyInt) Get() int { return int(m) }

type Arr [2]int

func (a Arr) Get() int { return a[0] + a[1] }

type Box struct{ IG } // struct embedding, but holding non-struct impls

func main() {
	var o IG = MyInt(5)
	g := o.Get
	println("int:", g())

	o = Arr{3, 4}
	g2 := o.Get
	println("arr:", g2())

	b := &Box{IG: MyInt(9)} // nested hop lands on a non-struct operand
	var o2 IG = b
	g3 := o2.Get
	println("boxed:", g3())
}

// Output:
// int: 5
// arr: 7
// boxed: 9
