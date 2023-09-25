package append

import (
	"gno.land/p/demo/ufmt"
)

type T struct{ i int }

var a []T

func Append(i int) {
	a = append(a, T{i: i})
}

func Pop() {
	a = append(a[:0], a[1:]...)
}

func Render(_ string) string {
	var s string
	for i := 0; i < len(a); i++ {
		s += ufmt.Sprintf("%d</br>", a[i].i)
	}
	return s
}
