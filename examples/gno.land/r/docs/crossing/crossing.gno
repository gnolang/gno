package crossing

import (
	"chain/runtime"
)

// This realm demonstrates the crossing mechanism in Gno

var owner address

func init() {
	owner = runtime.CurrentRealm().Address()
}

// PublicFunction can be called by anyone
func PublicFunction(_ realm) string {
	caller := runtime.PreviousRealm().Address()
	return "Called by: " + caller.String()
}

// InternalHelper is a helper function (no realm parameter)
// It can still be called from other realms but is typically
// used internally within this realm
func InternalHelper() string {
	return "I'm a helper function"
}

// GetOwner returns the owner address
func GetOwner() address {
	return owner
}

func Render(_ string) string {
	return `# Crossing in Gno

Crossing is how Gno tracks which realm is executing code and who called it.

## What is Crossing?

Every realm in Gno has its own isolated state. When Realm A calls a function
in Realm B, execution "crosses" into Realm B. The Gno runtime tracks this
call chain so each realm knows:

- **Who called me?** → ` + "`runtime.PreviousRealm()`" + `
- **Where am I?** → ` + "`runtime.CurrentRealm()`" + `
- **Who started the transaction?** → ` + "`runtime.OriginCaller()`" + `
- **Whose state is being modified?** → Always the current realm's state

This tracking is essential for authorization and security.

## The ` + "`realm`" + ` Parameter

Functions that modify realm state should include a ` + "`_ realm`" + ` parameter:

` + "```gno" + `
// This function crosses the realm boundary
func ModifyState(_ realm, newValue string) {
    state = newValue  // Modifies THIS realm's state
}
` + "```" + `

The ` + "`realm`" + ` parameter signals that this function:
1. Represents a realm boundary crossing
2. May modify this realm's state
3. Updates the call chain for runtime functions

## Crossing in Realms (/r/)

Realms (under ` + "`/r/`" + `) have their own isolated state:

` + "```gno" + `
// /r/counter/counter.gno
package counter

var count int  // This is /r/counter's state

func Increment(_ realm) {
    count++  // Always modifies /r/counter's state
}
` + "```" + `

When ` + "`/r/myapp`" + ` calls ` + "`/r/counter.Increment()`" + `:
- The count variable lives in ` + "`/r/counter`" + `'s state
- ` + "`runtime.CurrentRealm()`" + ` returns ` + "`/r/counter`" + `
- ` + "`runtime.PreviousRealm()`" + ` returns ` + "`/r/myapp`" + `

## The Call Chain

When User → Realm A → Realm B → Realm C:

` + "```" + `
 User Transaction                                            
                                                             
  User (EOA)                                                 
    │                                                        
    ▼ call                                                   
  ┌─────────────┐  OriginCaller() = User                     
  │   Realm A            │  PreviousRealm() = User (as realm)         
  │                      │  CurrentRealm() = Realm A                  
  └─────────────┘                                            
    │                                                        
    ▼ cross                                                  
  ┌─────────────┐  OriginCaller() = User                     
  │   Realm B            │  PreviousRealm() = Realm A                 
  │                      │  CurrentRealm() = Realm B                  
  └─────────────┘               
    │                                                                                  
    ▼ cross                                                  
  ┌─────────────┐  OriginCaller() = User                     
  │   Realm C            │  PreviousRealm() = Realm B                 
  │                      │  CurrentRealm() = Realm C                  
  └─────────────┘                                            
                                                             
` + "```" + `


## Crossing in Packages (/p/)

Packages (under ` + "`/p/`" + `) don't have their own persistent state. When a
realm imports a package, the package's variables are stored in that realm:

` + "```gno" + `
// /p/demo/counter/counter.gno
package counter

type Counter struct {
    value int
}

func New() *Counter {
    return &Counter{}
}

func (c *Counter) Increment() {
    c.value++
}
` + "```" + `

Each realm that imports and uses this package gets its own Counter instance
stored in its own state.

## Crossing in Tests

Tests can simulate crossing using testing utilities:

` + "```gno" + `
import "testing"

func TestCrossRealm(t *testing.T) {
    // Simulate a specific caller
    testing.SetRealm(testing.NewRealm("gno.land/r/caller"))
    
    // Now PreviousRealm() returns the simulated realm
    result := SomeFunction()
    
    // Or set the origin caller (EOA)
    testing.SetOriginCaller(testing.TestAddress("user1"))
}
` + "```" + `

## Security Implications

### Use PreviousRealm for Authorization

` + "```gno" + `
// SAFE: Checks immediate caller
func AdminOnly(_ realm) {
    if runtime.PreviousRealm().Address() != admin {
        panic("unauthorized")
    }
    // Admin action...
}
` + "```" + `

### Use OriginCaller for Logging

` + "```gno" + `
// OK for logging: Records the original user
func LogAction(_ realm) {
    user := runtime.OriginCaller()
    logs = append(logs, "Action by: " + user.String())
}
` + "```" + `

---

See also:
- [Common Mistakes](/r/docs/mistakes) - OriginCaller vs PreviousRealm
- [Safe Objects](/r/docs/safeobjects) - Exposing objects safely
`
}
