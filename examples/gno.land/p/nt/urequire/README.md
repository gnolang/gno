# `urequire` - Micro Require Assertions

A sister package to `uassert` that provides assertion functions which immediately fail the test (like `require` in testify). Unlike `uassert` which returns boolean values, `urequire` functions call `t.FailNow()` on assertion failure, stopping test execution immediately.

## Features

- **Immediate failure**: Test stops execution on first failed assertion
- **Same API as uassert**: Familiar function signatures
- **Built on uassert**: Leverages the same underlying assertion logic
- **Test helper support**: Proper test helper integration

## Usage

```go
import "gno.land/p/nt/urequire"

func TestCriticalFlow(t *testing.T) {
    // These must pass for test to continue
    config := loadConfig()
    urequire.NotNil(t, config) // Test stops here if config is nil
    
    // This line won't execute if config was nil
    urequire.Equal(t, "production", config.Environment)
    
    // Continue with test knowing config is valid
    result := processWithConfig(config)
    urequire.NoError(t, result.Error)
}
```

## When to Use urequire vs uassert

**Use `urequire` when:**
- Setup conditions must be met for test to continue
- Early failure saves time and provides clearer error messages
- Testing sequential operations where later steps depend on earlier ones

**Use `uassert` when:**
- You want to check multiple conditions and see all failures
- Testing independent assertions that don't affect each other
- Generating comprehensive test reports

## API

### Error Assertions
```go
func NoError(t uassert.TestingT, err error, msgs ...string)
func Error(t uassert.TestingT, err error, msgs ...string)
func ErrorContains(t uassert.TestingT, err error, contains string, msgs ...string)
```

### Equality Assertions
```go
func Equal(t uassert.TestingT, expected, actual any, msgs ...string)
func NotEqual(t uassert.TestingT, expected, actual any, msgs ...string)
```

### Boolean Assertions
```go
func True(t uassert.TestingT, value bool, msgs ...string)
func False(t uassert.TestingT, value bool, msgs ...string)
```

### Nil Assertions
```go
func Nil(t uassert.TestingT, object any, msgs ...string)
func NotNil(t uassert.TestingT, object any, msgs ...string)
```

### String Assertions
```go
func Contains(t uassert.TestingT, s, substr string, msgs ...string)
func NotContains(t uassert.TestingT, s, substr string, msgs ...string)
func Empty(t uassert.TestingT, object any, msgs ...string)
func NotEmpty(t uassert.TestingT, object any, msgs ...string)
```

### Panic Assertions
```go
func Panics(t uassert.TestingT, f func(), msgs ...string)
func NotPanics(t uassert.TestingT, f func(), msgs ...string)
```

## Example Comparison

```go
// Using uassert - continues even if assertions fail
func TestWithAssert(t *testing.T) {
    uassert.NotNil(t, config)      // Returns false but continues
    uassert.Equal(t, "prod", env)  // Still executes even if config was nil
    uassert.NoError(t, err)        // Still executes
    // All failures are reported
}

// Using urequire - stops on first failure
func TestWithRequire(t *testing.T) {
    urequire.NotNil(t, config)     // Test stops here if config is nil
    urequire.Equal(t, "prod", env) // Only executes if config is not nil
    urequire.NoError(t, err)       // Only executes if env matches
    // Test stops at first failure
}
```

## Dependencies

- `gno.land/p/nt/uassert` - Underlying assertion logic

This package is essential for writing tests where preconditions must be met before continuing with the test execution, providing cleaner failure modes and more focused error reporting.
