package tdd

// Calculator provides simple arithmetic operations
// This is designed to demonstrate TDD principles

var lastResult int64

func Add(a, b int64) int64 {
	lastResult = a + b
	return lastResult
}

func Subtract(a, b int64) int64 {
	lastResult = a - b
	return lastResult
}

func Multiply(a, b int64) int64 {
	lastResult = a * b
	return lastResult
}

func Divide(a, b int64) int64 {
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

First, write the test:

` + "```go" + `
// calculator_test.gno
func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
` + "```" + `

Then write the code:

` + "```go" + `
// calculator.gno
func Add(a, b int64) int64 {
    return a + b
}
` + "```" + `

## Table-Driven Tests

Test multiple cases efficiently:

` + "```go" + `
func TestAdd(t *testing.T) {
    tests := []struct {
        a, b, want int64
    }{
        {2, 3, 5},
        {-2, -3, -5},
        {0, 5, 5},
    }
    
    for _, tt := range tests {
        got := Add(tt.a, tt.b)
        if got != tt.want {
            t.Errorf("Add(%d, %d) = %d; want %d", 
                tt.a, tt.b, got, tt.want)
        }
    }
}
` + "```" + `

## Testing Panics

` + "```go" + `
func TestDivideByZero(t *testing.T) {
    defer func() {
        if r := recover(); r == nil {
            t.Error("expected panic")
        }
    }()
    Divide(10, 0)
}
` + "```" + `

## Testing with Addresses

` + "```go" + `
func TestWithCaller(t *testing.T) {
    addr := testing.TestAddress("user1")
    testing.SetOriginCaller(addr)
    
    result := SomeFunctionThatChecksCaller()
    // Assert...
}
` + "```" + `

## Testing with Coins

` + "```go" + `
func TestWithCoins(t *testing.T) {
    testing.SetOriginSend(chain.Coins{chain.NewCoin("ugnot", 1000)})
    
    result := SomePayableFunction()
    // Assert...
}
` + "```" + `

## Running Tests

` + "```bash" + `
gno test .              # Run all tests
gno test -v .           # Verbose output
gno test -run TestAdd . # Run specific test
` + "```" + `

[View source code](/r/docs/tdd$source)

---

See also:
- [gno test documentation](https://docs.gno.land/reference/gno-cli/gno-test)
- [Crossing](/r/docs/crossing) - How to test realm boundaries
- [Common Mistakes](/r/docs/mistakes) - Pitfalls to avoid
`
}
