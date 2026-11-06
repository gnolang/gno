package main

// doOpDefer located the deferred func by peeking past len(Call.Args)
// operands, but an embedded multi-value call leaves len(Args[0].Results) of
// them. `defer f(g())` therefore bound the defer to one of g's results: with
// a func-typed first result it deferred that closure instead of f, and with
// scalar results it panicked "defer called a nil function".
func g3() (int, int, int) { return 1, 2, 3 }

func f3(a, b, c int) { println("f3:", a, b, c) }

func gf() (func(), int) { return func() { println("closure") }, 7 }

func ff(fn func(), i int) { println("ff:", i); fn() }

// Variadic, method and builtin callees failed the same way.
type counter struct{ n int }

func (c counter) add(a, b int) { println("method:", c.n+a+b) }

func two() (int, int) { return 7, 8 }

func variadic(a int, rest ...int) { println("variadic:", a, len(rest), rest[0]) }

func main() {
	defer f3(g3())
	defer ff(gf())
	defer println(two())
	defer counter{n: 1}.add(two())
	defer variadic(two())
	println("main")
}

// Output:
// main
// variadic: 7 1 8
// method: 16
// 7 8
// ff: 7
// closure
// f3: 1 2 3
