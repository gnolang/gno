package composition

import (
	"gno.land/r/docs/composition/counter"
	"gno.land/r/docs/composition/logger"
)

// This realm demonstrates composition by combining functionality
// from two sub-realms: a counter and a logger.

// IncrementAndLog increments the counter and logs the action
func IncrementAndLog(cur realm) {
	counter.Increment(cross(cur))
	logger.Log(cross(cur), "Counter incremented to "+counter.GetCountStr())
}

// DecrementAndLog decrements the counter and logs the action
func DecrementAndLog(cur realm) {
	counter.Decrement(cross(cur))
	logger.Log(cross(cur), "Counter decremented to "+counter.GetCountStr())
}

// Reset resets the counter and clears logs
func Reset(cur realm) {
	counter.Reset(cross(cur))
	logger.Clear(cross(cur))
	logger.Log(cross(cur), "System reset")
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
    // Pass cross(cur) when calling external realms
    a.DoSomething(cross(cur))
    b.DoSomethingElse(cross(cur))
}
` + "```" + `

## Access Control Pattern

To prevent users from bypassing the composition realm and calling components directly,
each component checks that the caller is the authorized composition realm:

` + "```go" + `
// In component realm
const allowedCaller = "gno.land/r/my/composite"

func ModifyState(cur realm) {
    caller := cur.Previous()
    if caller.PkgPath() != allowedCaller {
        panic("unauthorized")
    }
    // ... modify state
}
` + "```" + `

**Alternative**: If components don't need their own persistent state, 
use pure packages (` + "`/p/...`" + `) instead of realms (` + "`/r/...`" + `). 
Packages are stateless libraries that can only be imported, so all state 
lives in the parent realm, eliminating bypass concerns entirely.

## This Example

This realm combines a counter and a logger. Each action increments or decrements
the counter while automatically logging the change.

The counter and logger can only be modified through this composition realm.
Direct calls to ` + "`counter.Increment()`" + ` or ` + "`logger.Log()`" + ` will fail with an authorization error.

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
