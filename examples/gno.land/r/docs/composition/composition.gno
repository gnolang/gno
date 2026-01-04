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

Allows you to build complex realms by combining simpler ones,
following the "separation of concerns" principle.

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


---

## Live Demo

- **Counter** ([source](/r/docs/composition/counter))
- **Logger** ([source](/r/docs/composition/logger))

### Current Counter Value

` + counter.RenderValue() + `

### Activity Log

` + logger.RenderLog() + `

### Actions

- [Increment and Log](/r/docs/composition$help&func=IncrementAndLog)
- [Decrement and Log](/r/docs/composition$help&func=DecrementAndLog)
- [Reset All](/r/docs/composition$help&func=Reset)
`
}
