# Gno Doctest

Gno Doctest executes fenced code blocks from Markdown files against
the Gno VM and verifies their output.

## Usage

```
gno doctest -path <markdown_file> [-run <regex>]
```

- `<markdown_file>`: Path to the Markdown source.
- `<regex>`: Optional Go-style pattern; only blocks whose name matches
  are executed.

## Recognized fenced-code languages

A code block is treated as Gno when its language tag is `gno` or
includes `gnodoctest` (e.g. `go,gnodoctest`, useful when you want
GitHub to render Go syntax highlighting for the snippet).

## Directives

Doctest reuses [filetest](../test/filetest.go)'s directive grammar.
PascalCase keys (`Output`, `Error`) start a multi-line section
captured from the following `// ...` comment lines until the next
section, a bare `//`, or any non-comment line. ALLCAPS keys are
single-line: `KEY:` is a flag, `KEY: value` carries a value.

| Directive | Form | Purpose |
| --------- | ---- | ------- |
| `Output` | multi-line | Expected stdout. |
| `Error`  | multi-line | Expected error/panic text. |
| `NAME`   | `NAME: <name>` | Block name used by `-run`. Defaults to `block_<index>`. |
| `IGNORE` | `IGNORE:` | Skip execution of this block. |
| `SHOULD_PANIC` | `SHOULD_PANIC:` or `SHOULD_PANIC: <msg>` | Block must panic. With a value, the panic must contain that substring. |

Output values prefixed with `regex:` are matched as a regular
expression instead of a literal string.

## Examples

Basic execution with expected output:

````
```gno
package main

func main() {
    println("Hello, World!")
}

// Output:
// Hello, World!
```
````

Named block, executable via `-run hello`:

````
```gno
// NAME: hello
package main

func main() { println("hi") }

// Output:
// hi
```
````

Skipped block:

````
```gno
// IGNORE:
package main

func main() {
    panic("never runs")
}
```
````

Panic expectation with message:

````
```gno
// SHOULD_PANIC: index out of range
package main

func main() {
    a := []int{1, 2, 3}
    println(a[5])
}
```
````
