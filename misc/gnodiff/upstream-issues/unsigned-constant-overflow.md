# Draft: do not open as a new issue

This is [gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392), which is still open on upstream master. Add the row below to that issue. A suggested comment is at the end.

## Title

gnovm: typed constant `uint64(8) - uint64(40)` wraps instead of failing

## Summary

In Go, a constant expression with a typed operand must fit in that type. `uint64(8) - uint64(40)` is the constant `-32` and does not fit in `uint64`, so gc rejects it. Gno's preprocessor folds the same expression with runtime arithmetic, which wraps, and `gno run` prints `18446744073709551584` (`2^64 - 32`).

`gno lint` and `gno test` already reject it, because they run go/types. The bug is the VM fold used by `gno run`, and filetest execution, which still runs after recording the typecheck error.

The same subtraction through variables is not a bug. gc and `gno run` both print `18446744073709551584`. Runtime unsigned wrap should stay.

Not described as intentional. #6392 already covers typed constant `+`, `-`, `*`, and negation, including `const d uint8 = 0; d - 1`. This expression is one more row of that table.

## Minimal repro

```go
package main

func main() {
	println(uint64(8) - uint64(40))
}
```

Control, which both toolchains accept:

```go
package main

func main() {
	var a uint64 = 8
	var b uint64 = 40
	println(a - b)
}
```

## Expected

Reject the constant form:

```
./main.go:4:10: uint64(8) - uint64(40) (constant -32 of type uint64) overflows uint64
```

The variable form prints `18446744073709551584`.

## Actual

Validated on upstream `master` `cf8ff126706cc9ba2d226bc256b3060dfa7731e2`, Go `go1.25.9 linux/amd64`, `gno` built from that commit. `GNOROOT` was the clone root.

`gno run .` on the constant form: stdout `18446744073709551584`, exit 0.

`gno run .` on the variable form: the same number, exit 0. `go run` on the variable form prints the same number.

`gno lint .` and `gno test .` on package `gno.land/p/demo/overflowbug` and realm `gno.land/r/demo/overflowbug`: exit 1,

```
uint64(8) - uint64(40) (constant -32 of type uint64) overflows uint64 (code=gnoTypeCheckError)
```

Filetest `TestFiles/zz_unsigned_sub_repro.gno` with that number as `// Output:`: the output matches, and `TypeCheckError` is

```
main/zz_unsigned_sub_repro.gno:4:10: uint64(8) - uint64(40) (constant -32 of type uint64) overflows uint64
```

Execution is not skipped ([gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601)).

## Related

- [gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392) (open): the general bug. Do not file a second issue.
- [gnolang/gno#6409](https://github.com/gnolang/gno/pull/6409) (open): fold constants with `go/constant`. Not merged on `cf8ff126`.
- [gnolang/gno#6390](https://github.com/gnolang/gno/pull/6390) (open): untyped rune constants only. #6392 says that change leaves typed constants alone.
- [gnolang/gno#6389](https://github.com/gnolang/gno/issues/6389) (open): untyped rune constant arithmetic wraps at int32. Adjacent, not this expression.
- [gnolang/gno#6330](https://github.com/gnolang/gno/issues/6330) (open): out-of-range constant int-to-string. Different conversion.

## Comment to paste on #6392

One more row, reproduced on master `cf8ff126706cc9ba2d226bc256b3060dfa7731e2` with Go 1.25.9. `gno run` prints `18446744073709551584`. `gno lint` and `gno test` (package and realm) reject with `overflows uint64`. A filetest records that `TypeCheckError` and still prints the wrapped value. The non-constant form `var a, b uint64 = 8, 40; println(a - b)` matches gc, so this is only the constant fold.
