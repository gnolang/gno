package tdd

// Calculator provides simple arithmetic operations
// This is designed to demonstrate TDD principles

var lastResult int64

func Add(_ realm, a, b int64) int64 {
	lastResult = a + b
	return lastResult
}

func Subtract(_ realm, a, b int64) int64 {
	lastResult = a - b
	return lastResult
}

func Multiply(_ realm, a, b int64) int64 {
	lastResult = a * b
	return lastResult
}

func Divide(_ realm, a, b int64) int64 {
	if b == 0 {
		panic("division by zero")
	}
	lastResult = a / b
	return lastResult
}

func GetLastResult() int64 {
	return lastResult
}

func Render(_ string) string {
	return `# Test-Driven Development (TDD) in Gno

TDD is writing tests BEFORE writing code.

## The TDD Cycle

1. **Red** - Write a failing test
2. **Green** - Write minimal code to pass
3. **Refactor** - Improve while keeping tests green

## Quick Example

First, write the test. A function that changes realm state takes a cur realm
parameter, so the test takes one too and calls it across the realm boundary
with cross(cur):

` + "```go" + `
import "testing"

// calculator_test.gno
func TestAdd(cur realm, t *testing.T) {
    result := Add(cross(cur), 2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
` + "```" + `

Then write the code:

` + "```go" + `
// calculator.gno
func Add(_ realm, a, b int64) int64 {
    return a + b
}
` + "```" + `

## Table-Driven Tests

Test multiple cases efficiently:

` + "```go" + `
func TestAdd(cur realm, t *testing.T) {
    tests := []struct {
        a, b, want int64
    }{
        {2, 3, 5},
        {-2, -3, -5},
        {0, 5, 5},
    }

    for _, tt := range tests {
        got := Add(cross(cur), tt.a, tt.b)
        if got != tt.want {
            t.Errorf("Add(%d, %d) = %d; want %d",
                tt.a, tt.b, got, tt.want)
        }
    }
}
` + "```" + `

## Testing Panics

A panic that crosses a realm boundary aborts the program, so it can't be caught
with recover. Use the revive builtin to assert that a cross-call panics:

` + "```go" + `
func TestDivideByZero(cur realm, t *testing.T) {
    rec := revive(func() {
        Divide(cross(cur), 10, 0)
    })
    if rec == nil {
        t.Error("expected panic")
    }
}
` + "```" + `

## Testing with Addresses

` + "```go" + `
import "gno.land/p/nt/testutils/v0"

func TestWithCaller(cur realm, t *testing.T) {
    addr := testutils.TestAddress("user1")
    testing.SetOriginCaller(addr)

    result := SomeFunctionThatChecksCaller(cross(cur))
    // Assert...
}
` + "```" + `

## Testing with Coins

` + "```go" + `
func TestWithCoins(cur realm, t *testing.T) {
    testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", 1000)})

    result := SomePayableFunction(cross(cur))
    // Assert...
}
` + "```" + `

## Running Tests

` + "```bash" + `
gno test .              # Run all tests
gno test -v .           # Verbose output
gno test -run TestAdd . # Run specific test
` + "```" + `

---

### See also:
- [gno test documentation](https://docs.gno.land/reference/gno-cli/gno-test)
- [Crossing](/r/docs/crossing) - How to test realm boundaries
- [Common Mistakes](/r/docs/mistakes) - Pitfalls to avoid
`
}
