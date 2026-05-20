# Running & Testing Gno code

`gno test` runs `_test.gno` files for a Gno package; `gno run` evaluates
expressions against package code. Both use a mocked GnoVM environment — no
real chain, state changes are in-memory only.

## Prerequisites

`gno` installed — see [Installation](../builders/install.md). Examples below
use the `Counter` realm from
[Getting started](../builders/getting-started.md), with `counter.gno` and
`counter_test.gno` in the package directory.

## `gno test`

From inside the package directory:

```
$ gno test .
ok      .       0.81s
```

Add `-v` for verbose output:

```
$ gno test . -v
=== RUN   TestIncrement
--- PASS: TestIncrement (0.00s)
ok      .       0.81s
```

Other flags cover test timeouts and performance checks — see `gno test --help`.

## `gno run`

`gno run` evaluates an expression against package code. It's a program
runner, not a REPL, so return values aren't printed automatically. Wrap
in `println()`:

```
$ gno run -expr "println(Increment(42))"
42
```

Pass `-debug` to start the GnoVM debugger; see this
[blog post](https://gno.land/r/gnoland/blog:p/gno-debugger).

## Filetests

Filetests are golden tests typically used to test realms. They execute a `main`
function and compare actual output against expected output written as comment
directives at the bottom of the file.

Filetests use the `*_filetest.gno` suffix and are placed in a `filetests/`
subdirectory of the realm package.

:::warning Stability notice
Filetests are primarily intended as an internal tool. Their API and behavior
are not guaranteed to be as stable as standard `gno test` testing.
:::

### Example

```go
// PKGPATH: gno.land/r/demo/counter_test
// SEND: 1000000ugnot
package counter_test

import "gno.land/r/demo/counter"

func main() {
	counter.Increment(cross)
	println(counter.Render(""))
}

// Output:
// 1
```

### Running filetests

```bash
# Only run the filetest for a package (from the package directory)
gno test -run "_filetest.gno" .
# Update expected values when output intentionally changes
gno test --update-golden-tests .
```

### Directives

**Input directives** are single-line comments at the top of the file:

| Directive  | Description                          | Default |
|------------|--------------------------------------|---------|
| `PKGPATH`  | Package path. Use `r/` for realms.   | `main`  |
| `MAXALLOC` | Max memory allocation in bytes.      | `0`     |
| `SEND`     | Coins sent with the transaction.     | (none)  |

**Output directives** are multi-line comments at the bottom:

| Directive        | Matches                                       |
|------------------|-----------------------------------------------|
| `Output`         | Standard output.                              |
| `Error`          | Panic or error message.                       |
| `Realm`          | Realm state change operations.                |
| `Events`         | Emitted events (JSON).                        |
| `Preprocessed`   | Preprocessed AST.                             |
| `Stacktrace`     | Gno stacktrace on panic.                      |
| `Gas`            | Gas consumed.                                 |
| `Storage`        | Realm storage size diff.                      |
| `TypeCheckError` | Go type-checker error.                        |

:::info Pure package imports
Imports of pure packages are processed separately. If a pure package contains a
line like `println(1)`, its output cannot be checked by an `// Output:` directive.
:::
