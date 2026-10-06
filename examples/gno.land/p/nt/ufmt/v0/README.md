> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `ufmt` - String formatting

Gno port of a subset of Go's `fmt` package (micro-fmt). Provides `Printf`, `Sprintf`, `Errorf` and friends for formatting strings with verb-based templates.

## Usage

```go
import "gno.land/p/nt/ufmt/v0"

s := ufmt.Sprintf("hello %s, you are %d years old", "alice", 30)
// "hello alice, you are 30 years old"

err := ufmt.Errorf("invalid id: %q", input)

var buf bytes.Buffer
ufmt.Fprintf(&buf, "balance: %d", amount)

line := ufmt.Sprintln("token", symbol, "transferred") // adds spaces + newline
```

## API

```go
// Format and return a string.
func Sprint(a ...any) string
func Sprintf(format string, a ...any) string
func Sprintln(a ...any) string

// Format and write to an io.Writer.
func Fprint(w io.Writer, a ...any) (n int, err error)
func Fprintf(w io.Writer, format string, a ...any) (n int, err error)
func Fprintln(w io.Writer, a ...any) (n int, err error)

// Format and print to standard output.
func Print(a ...any) (n int, err error)
func Printf(format string, a ...any) (n int, err error)
func Println(a ...any) (n int, err error)

// Format and append to a byte slice.
func Append(b []byte, a ...any) []byte
func Appendf(b []byte, format string, a ...any) []byte
func Appendln(b []byte, a ...any) []byte

// Format and return an error.
func Errorf(format string, args ...any) error
```

## Supported verbs

| Verb | Meaning                                                                |
|------|------------------------------------------------------------------------|
| `%s` | String. Uses `String()` or `Error()` if implemented.                   |
| `%d` | Integer (signed and unsigned, all widths).                             |
| `%c` | Unicode character from rune/int code point.                            |
| `%t` | Boolean: `true` or `false`.                                            |
| `%q` | Double-quoted, escaped string.                                         |
| `%x` | Hexadecimal (uint8 only).                                              |
| `%f` / `%F` | Decimal float; default precision 6.                             |
| `%e` / `%E` | Scientific notation float; default precision 2.                 |
| `%g` / `%G` | Float, compact representation.                                  |
| `%T` | Type name of the argument (basic types only).                          |
| `%v` | Default representation appropriate for the value's type.               |
| `%%` | Literal `%`.                                                           |

Width (`%5s`) and precision (`%.2f`) are supported for the relevant verbs.

## Notes

- Verb/type mismatches produce `%!verb(type=value)` strings, matching Go's `fmt` behaviour.
- Missing or extra arguments panic.
- Not supported: `%b`, `%o`, `%U`, `%p`, `%+v`, `%#v`, argument indexing, flags like `-`, `+`, `#`, `0`.
- `Print*` writes via the built-in `print` (stdout substitute) until `os.Stdout` is available.
