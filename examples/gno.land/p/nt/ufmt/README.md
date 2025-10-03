# `ufmt` - Micro Format Utilities

A lightweight formatting library similar to Go's `fmt` package, providing string formatting functions for Gno programs. This is a subset implementation (hence "µfmt" - micro fmt) focusing on the most commonly used formatting operations.

## Features

- **Printf-style formatting**: Support for format verbs and placeholders
- **Multiple output targets**: Print to strings, writers, or create errors
- **Type-safe formatting**: Proper handling of different Go types
- **Memory efficient**: Optimized buffer management

## Usage

```go
import "gno.land/p/nt/ufmt"

// Format to string
str := ufmt.Sprintf("Hello %s, you are %d years old", "Alice", 30)
// "Hello Alice, you are 30 years old"

// Print with formatting
ufmt.Printf("User: %s, Score: %d\n", username, score)

// Create formatted error
err := ufmt.Errorf("failed to process user %s: %v", username, originalErr)

// Write to io.Writer
var buf strings.Builder
ufmt.Fprintf(&buf, "Data: %v", data)
```

## Supported Format Verbs

- `%s` - String representation
- `%d` - Decimal integer
- `%v` - Default format for any value
- `%t` - Boolean (true/false)
- `%x` - Hexadecimal lowercase
- `%X` - Hexadecimal uppercase
- `%o` - Octal
- `%b` - Binary
- `%f` - Floating point
- `%c` - Character (rune)
- `%q` - Quoted string
- `%%` - Literal percent sign

## API

```go
// Format to string
func Sprintf(format string, args ...interface{}) string

// Print to stdout
func Printf(format string, args ...interface{})

// Write to io.Writer
func Fprintf(w io.Writer, format string, args ...interface{}) (int, error)

// Create formatted error
func Errorf(format string, args ...interface{}) error
```

## Examples

```go
// Basic formatting
name := "Gno"
version := 1.0
ufmt.Printf("Welcome to %s v%.1f\n", name, version)

// Number formatting
num := 255
ufmt.Printf("Decimal: %d, Hex: %x, Binary: %b\n", num, num, num)

// Boolean and character
isActive := true
initial := 'G'
ufmt.Printf("Active: %t, Initial: %c\n", isActive, initial)

// Error creation
if err != nil {
    return ufmt.Errorf("operation failed: %v", err)
}
```

This package is essential for string formatting and debugging in Gno programs, providing familiar Printf-style functionality in a lightweight implementation.
