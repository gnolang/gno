package composition

import (
	"gno.land/r/docs/composition/counter"
	"gno.land/r/docs/composition/logger"
)

// This realm demonstrates composition by combining functionality
// from two sub-realms: a counter and a logger.

// IncrementAndLog increments the counter and logs the action
func IncrementAndLog(cur realm) {
	counter.Increment(cur)
	logger.Log(cur, "Counter incremented to "+counter.GetCountStr())
}

// DecrementAndLog decrements the counter and logs the action
func DecrementAndLog(cur realm) {
	counter.Decrement(cur)
	logger.Log(cur, "Counter decremented to "+counter.GetCountStr())
}

// Reset resets the counter and clears logs
func Reset(cur realm) {
	counter.Reset(cur)
	logger.Clear(cur)
	logger.Log(cur, "System reset")
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

func ComposedAction(cur realm) {
    // Call functions from component A
    a.DoSomething(cur)
    
    // Call functions from component B
    b.DoSomethingElse(cur)
}
` + "```" + `

## This Example

This realm combines a counter and a logger. Each action increments or decrements
the counter while automatically logging the change.

**Key concept**: When calling functions in other realms that modify state, 
you must pass the ` + "`realm`" + ` parameter to enable cross-realm calls.

## Components

- **Counter** ([source](/r/docs/composition/counter))
- **Logger** ([source](/r/docs/composition/logger))

## Live Demo

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
