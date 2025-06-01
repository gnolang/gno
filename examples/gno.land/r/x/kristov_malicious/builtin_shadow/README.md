Redefining or "shadowing" built-ins can facilitate serruptitious code. Consider the following example:

```go
package main

func main() {
	panic("foo")
}

func panic(s string) {
	println("bar")
}
```

In Go, results in printing:
```
bar
```

Gno should deliberately allow this or disallow this and document the outcome.

This was addressed in the following PR: https://github.com/gnolang/gno/issues/1779
