# Running & Testing Gno code

The `gno` binary runs and tests Gno code locally: `gno test` for unit tests,
`gno run` for evaluating one-off expressions, and filetests for golden tests of
realms.

All of them run against a mocked GnoVM, so there is no real chain and any state
changes stay in memory for that one command. Imports resolve from your local
gno installation rather than a chain, so tests see the same standard library and
examples packages you have on disk.

## Prerequisites

`gno` installed. See [Installation](../builders/install.md). The examples below
build on the counter realm from
[Getting started](../builders/getting-started.md): the `myrealm` package, with
`myrealm.gno` and `myrealm_test.gno` in the package directory.

## `gno test`

`gno test` runs a package's `_test.gno` files, much like `go test`. From inside
the package directory:

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

Other flags cover test timeouts and performance checks. See `gno test --help`.

## `gno run`

`gno run` evaluates an expression against your package code, a quick way to
check a function during development without deploying. It works with pure
packages and plain, non-crossing functions. To exercise realm functions that
take a `realm` argument, use `gno test` or a filetest instead.

It's a program runner, not a REPL, so return values aren't printed
automatically. Wrap the expression in `println()`:

```
$ gno run -expr "println(Add(2, 3))" .
5
```

Pass `-debug` to start the GnoVM debugger. See this
[blog post](https://gno.land/r/gnoland/blog:p/gno-debugger).

## Example tests

`gno test` also supports example tests, [similar to Go](https://go.dev/blog/examples). An
example test function takes no arguments and begins with the word `Example`. Like the test shown
above, it must be in a file ending in `_test.gno`.
The function prints output which is compared to the expected output in the `// Output:` comment.

To try it, create a file `example_test.gno` which checks the expected value of the `Render` function:
```
touch example_test.gno
```

`example_test.gno`:
```go
package myrealm

import (
	"fmt"
)

func ExampleRender() {
	count = 10
	fmt.Println(Render(""))
	// Output:
	// Count: 10
}
```

:::warning Reserved function name
Your test file can have local helper functions, but `init()` is reserved for other types of tests.
Use something like `initialize()` instead.
:::

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

## Integration tests

Unit tests and filetests run in the GnoVM alone. Behaviour that only shows up on
a real chain — gas and storage fees, emitted events, multi-transaction flows,
node restarts, `gnokey` output — is covered by
[testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
`.txtar` archives, which boot a node and drive it through `gnokey`.

Place a `.txtar` that is about a single package **next to that package**, in the
same directory as its `.gno` files; every `*.txtar` under `examples/` is
discovered automatically. Scripts that exercise the VM, the node or `gnokey`
itself rather than one package live in `gno.land/pkg/integration/testdata/`
instead. Both are run by the same test:

```bash
# every integration test
go test ./gno.land/pkg/integration/ -run TestTestdata -v
# just one, wherever it lives
go test ./gno.land/pkg/integration/ -run TestTestdata/wugnot -v
```

Two things to know:

- The subtest name is the file's base name alone, so base names must be unique
  across both locations — prefix them with the package name
  (`commondao_council.txtar`). A collision fails the suite.
- A `.txtar` is not part of the package. It is skipped when the directory is
  read into a mem-package, so it never ends up on-chain.
