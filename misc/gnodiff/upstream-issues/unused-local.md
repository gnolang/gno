# Draft: do not open as a new issue

Recommended action: comment on [gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426), and mention [gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601). Details of why are in [validation.md](validation.md).

If a maintainer wants a separate ticket, this is the text.

## Title

gnovm: `gno run` executes a local that is declared and not used

## Summary

`x := 1` with `x` never read is a compile error in Go. `gno lint` and `gno test` already reject it through go/types. `gno run` does not: it prints `2` and exits 0. A `gnovm/tests/files` filetest records the go/types error and then runs the function anyway.

This is not "the type checker cannot see unused locals." The checker used by lint and test sees this one. The execution path used by `gno run` does not fail on that result.

Not described as intentional in `docs/resources/go-gno-compatibility.md`.

## Minimal repro

```go
package main

func main() {
	x := 1
	println(2)
}
```

## Expected

Reject the program. gc says:

```
./main.go:4:2: x declared and not used
```

go/types, which `gno lint` prints, says:

```
main.gno:4:2: declared and not used: x
```

Either rejection is the Go behavior. The program must not run.

## Actual

Validated on upstream `master` `cf8ff126706cc9ba2d226bc256b3060dfa7731e2`, Go `go1.25.9 linux/amd64`, `gno` built from that commit (`gno version` prints `develop`). `GNOROOT` was the clone root.

`gno run .` from a directory containing only `main.gno`: stdout `2`, exit 0.

`gno lint .` on package `gno.land/p/demo/unusedbug` and on realm `gno.land/r/demo/unusedbug`: exit 1, `declared and not used: x (code=gnoTypeCheckError)`.

`gno test .` on both of those modules: the same typecheck error, `FAIL`, the test body does not pass.

Filetest `TestFiles/zz_unused_local_repro.gno` with an `// Output:` of `2`: the output matches, and the run also reports

```
main/zz_unused_local_repro.gno:4:2: declared and not used: x
```

as `TypeCheckError`. Execution is not skipped. That second half is [gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601).

## Related

- [gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426) (open): unused values are not rejected and `gno run` executes the program. Closest report. The example there is a bare `1` expression, not `x := 1`.
- [gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601) (open): filetests keep executing after `TypeCheckError`. The unused-local filetest does that.
- [gnolang/gno#3965](https://github.com/gnolang/gno/issues/3965) (closed, `not_planned`): inner `pid := 42` unread, outer `pid` prints 0. A different program.
