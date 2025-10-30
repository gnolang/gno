# `uassert` - Micro Assertion Library

A lightweight testing assertion library adapted from testify/assert, providing essential assertion functions for unit testing in Gno programs.

## Features

- **Familiar API**: Similar to popular Go testing libraries
- **Comprehensive assertions**: Support for errors, equality, comparisons, and more
- **Clear failure messages**: Detailed output when assertions fail
- **Diff support**: Visual differences for complex comparisons
- **Helper integration**: Proper test helper support

## Usage

```go
import "gno.land/p/nt/uassert"

func TestMyFunction(t *testing.T) {
    // Error assertions
    err := someFunction()
    uassert.NoError(t, err)
    
    // Equality assertions
    result := calculate(5, 3)
    uassert.Equal(t, 8, result)
    uassert.NotEqual(t, 0, result)
    
    // Boolean assertions
    uassert.True(t, isValid)
    uassert.False(t, isEmpty)
    
    // Nil assertions
    uassert.Nil(t, nilValue)
    uassert.NotNil(t, notNilValue)
    
    // String assertions
    uassert.Contains(t, "hello world", "world")
    uassert.Empty(t, "")
    uassert.NotEmpty(t, "content")
    
    // Panic assertions
    uassert.Panics(t, func() {
        panicFunction()
    })
    
    uassert.NotPanics(t, func() {
        safeFunction()
    })
}
```

## API

### Error Assertions
```go
func NoError(t TestingT, err error, msgs ...string) bool
func Error(t TestingT, err error, msgs ...string) bool
```

### Equality Assertions
```go
func Equal(t TestingT, expected, actual any, msgs ...string) bool
func NotEqual(t TestingT, expected, actual any, msgs ...string) bool
```

### Boolean Assertions
```go
func True(t TestingT, value bool, msgs ...string) bool
func False(t TestingT, value bool, msgs ...string) bool
```

### Nil Assertions
```go
func Nil(t TestingT, object any, msgs ...string) bool
func NotNil(t TestingT, object any, msgs ...string) bool
```

### String Assertions
```go
func Contains(t TestingT, s, substr string, msgs ...string) bool
func NotContains(t TestingT, s, substr string, msgs ...string) bool
func Empty(t TestingT, object any, msgs ...string) bool
func NotEmpty(t TestingT, object any, msgs ...string) bool
```

### Panic Assertions
```go
func Panics(t TestingT, f func(), msgs ...string) bool
func NotPanics(t TestingT, f func(), msgs ...string) bool
```

### Comparison Assertions
```go
func Greater(t TestingT, e1, e2 any, msgs ...string) bool
func GreaterOrEqual(t TestingT, e1, e2 any, msgs ...string) bool
func Less(t TestingT, e1, e2 any, msgs ...string) bool
func LessOrEqual(t TestingT, e1, e2 any, msgs ...string) bool
```

## Custom Messages

All assertion functions accept optional custom messages:

```go
uassert.Equal(t, expected, actual, "Values should be equal after calculation")
uassert.NoError(t, err, "Database connection should succeed")
```

## Interface

```go
type TestingT interface {
    Helper()
    Error(args ...interface{})
}
```

This interface is compatible with Go's `*testing.T` and similar testing frameworks.

## Dependencies

- `gno.land/p/nt/ufmt` - For formatted output
- `gno.land/p/onbloc/diff` - For showing differences in failed assertions

This package is essential for writing comprehensive unit tests in Gno, providing the familiar assertion patterns that Go developers expect.
