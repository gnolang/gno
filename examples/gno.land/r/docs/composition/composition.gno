package composition

import (
	"gno.land/r/docs/composition/counter"
	"gno.land/r/docs/composition/logger"
)

// This realm demonstrates composition by combining functionality
// from two sub-realms: a counter and a logger.

// IncrementAndLog increments the counter and logs the action
func IncrementAndLog(_ realm) {
	counter.Increment()
	logger.Log("Counter incremented to " + counter.GetCountStr())
}

// DecrementAndLog decrements the counter and logs the action
func DecrementAndLog(_ realm) {
	counter.Decrement()
	logger.Log("Counter decremented to " + counter.GetCountStr())
}

// Reset resets the counter and clears logs
func Reset(_ realm) {
	counter.Reset()
	logger.Clear()
	logger.Log("System reset")
}

func Render(_ string) string {
	return `# Realm Composition

This example demonstrates how to compose functionality from multiple realms.

## What is Composition?

Composition in Gno allows you to build complex functionality by combining
simpler, focused realms. This follows the principle of "separation of concerns"
and makes code more maintainable and reusable.

## This Example

This realm combines two sub-realms:
- **Counter** ([source](/r/docs/composition/counter)) - Manages a simple counter
- **Logger** ([source](/r/docs/composition/logger)) - Maintains an activity log

By composing these together, we get a logged counter without either realm
knowing about the other's implementation details.

Composition enables reusability, testability, maintainability, and enforces single responsibility.

## Live Demo

### Current Counter Value

` + counter.RenderValue() + `

### Activity Log

` + logger.RenderLog() + `

### Actions

- [Increment and Log](/r/docs/composition$help&func=IncrementAndLog)
- [Decrement and Log](/r/docs/composition$help&func=DecrementAndLog)
- [Reset All](/r/docs/composition$help&func=Reset)

---

## Code Pattern

` + "```go" + `
package mycomposite

import (
    "gno.land/r/component/a"
    "gno.land/r/component/b"
)

func ComposedAction(_ realm) {
    // Call functions from component A
    a.DoSomething()
    
    // Call functions from component B
    b.DoSomethingElse()
}
` + "```" + `

This pattern allows you to build sophisticated applications from simple building blocks.
`
}
