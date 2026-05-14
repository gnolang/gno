// Verifies the compile-only path: a .go file declaring a non-main
// package with no `func main()` passes when Gno's preprocess and
// go/types both accept the declarations. Exercises the
// PKGPATH+synthetic-main rescue. gc's `// compile` directive (when
// present on corpus files) is intentionally treated as a plain
// comment — only the file's actual behavior under Gno is checked.

package p

type T struct {
	X, Y int
}

func F(t T) int {
	return t.X + t.Y
}
