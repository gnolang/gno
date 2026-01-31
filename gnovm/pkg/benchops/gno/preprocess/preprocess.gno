package preprocess

import (
	"strconv"
)

type MyStruct struct {
	A int
	B string
}

type MyIface interface {
	M() int
}

type MyArray [2]int
type MySlice []int
type MyMap map[string]int
type MyChan chan int
type MyFunc func(int) int

type MyInterface interface{
	N() string
}

var globalVar int
const globalConst = 1

func (s MyStruct) M() int { return s.A }

func RunAll() int {
	var x int = 1
	y := 2
	x = y + globalConst

	neg := -x
	ok := !false

	ms := MyStruct{A: x, B: "b"}
	m := map[string]int{"a": x}
	s := []int{1, 2, 3}
	a := [2]int{1, 2}

	_ = m["a"]
	_ = s[1:2]
	_ = a[1]

	_ = ms.A
	_ = strconv.Itoa(x)

	px := &x
	_ = *px

	f := func(n int) int { return n + 1 }
	_ = f(3)

	var iv interface{} = x
	_ = iv.(int)

	if z := x + y; z > 0 {
		x = z
	} else if z == 0 {
		x = 0
	} else {
		x = -1
	}

	for i := 0; i < 2; i++ {
		if i == 1 {
			continue
		}
		x += i
	}

	for i, v := range s {
		x += i + v
	}

	switch x {
	case 1:
		x++
		fallthrough
	case 2:
		x++
	default:
		x++
	}

	switch v := iv.(type) {
	case int:
		x += v
	default:
		x += 0
	}

Loop:
	for i := 0; i < 3; i++ {
		if i == 2 {
			break Loop
		}
	}

	if x > 0 {
		goto done
	}
	x = 0

done:
	globalVar = x
	return x + neg + btoi(ok)
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
