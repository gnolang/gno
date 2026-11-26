// Package t76other is another realm with a plain (non-crossing) function.
package t76other

func Each(xs []int, f func(int)) {
	for _, x := range xs {
		f(x)
	}
}
