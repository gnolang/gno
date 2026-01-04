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

Crossing is the mechanism by which Gno tracks realm boundaries, determining which realm's state is currently being modified.

## What is Crossing?

Every realm in Gno has its own isolated state. When Realm A wants to execute
a function in Realm B's scope, execution "crosses" into Realm B. The Gno
runtime tracks this call chain so each realm knows:

- **Who started the transaction?** → ` + "`runtime.OriginCaller()`" + `
- **Who called me?** → ` + "`runtime.PreviousRealm()`" + `
- **Where am I?** → ` + "`runtime.CurrentRealm()`" + `
- **Whose state is being modified?** → Always the current realm's state

To define a function as crossing, you must include a ` + "`_ realm`" + ` parameter:

` + "```go" + `
func ModifyState(_ realm, newValue string) {
    state = newValue  // Modifies THIS realm's state
}
` + "```" + `

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

Unlike realms, packages (under ` + "`/p/`" + `) don't have their own persistent state.
When a realm imports a package, any variables declared by that package are
stored within the realm's state. This means each realm gets its
own isolated instance of the package's data.

## Crossing in [Tests](https://docs.gno.land/resources/gno-stdlibs/#testing)

` + "```go" + `
func TestWithCaller(t *testing.T) {
    // Simulate PreviousRealm()
    testing.SetRealm(testing.NewRealm("gno.land/r/caller"))
    
    // Simulate OriginCaller()
    testing.SetOriginCaller(testing.TestAddress("user1"))
}
` + "```" + `

---

### See also:
- [Common Mistakes](/r/docs/mistakes) - OriginCaller vs PreviousRealm
- [Safe Objects](/r/docs/safeobjects) - Exposing objects safely
`
}
