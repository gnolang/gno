// Package crossing explains Gno's crossing mechanism.
// Crossing tracks which realm is acting, and cross-calls move it from one realm to another.
package crossing

import (
	"chain/runtime/unsafe"
)

var owner address

func init(cur realm) {
	// init runs automatically only once on deployment. It can take a "cur realm"
	// handle, so "owner" is set from cur.Address(), the address of the realm
	// being deployed.
	owner = cur.Address()
}

// PublicFunction's first parameter is a "cur realm", so the call crosses into
// this realm. cur.Previous() is the idiomatic way to read the caller, but only
// after cur.IsCurrent() confirms the handle is this frame's live realm.
func PublicFunction(cur realm) string {
	if !cur.IsCurrent() {
		panic("invalid realm")
	}
	caller := cur.Previous().Address()
	return "Called by: " + caller.String()
}

// InternalHelper has no realm parameter, so the call does not cross. It runs
// in the caller's context, like a pure package helper.
func InternalHelper() string {
	caller := unsafe.CurrentRealm().Address()
	return "I'm a helper function called by: " + caller.String()
}

// InternalHelperWithRealm takes a realm but doesn't cross: only a first-position
// realm crosses, so rlm is a plain value here. It lets shared code that can't be
// a crossing function (e.g. a /p/ package) receive and verify the acting realm.
func InternalHelperWithRealm(_ int, rlm realm) string {
	// The call didn't cross, so IsCurrent() holds for the caller's live realm
	// and rejects any other.
	if !rlm.IsCurrent() {
		panic("rlm is not the current realm")
	}

	return "I'm a helper function called by: " + rlm.Address().String()
}

// GetOwner returns the owner address
func GetOwner() address {
	return owner
}

func Render(_ string) string {
	return `# Crossing in Gno

Crossing is how Gno tracks which realm is acting at any moment. Every realm is a
smart contract with its own storage, and a realm can only write to its own. When
one realm calls into another, the acting realm switches to the one being called.
Gno follows that switch so each change is saved to the realm that owns it.

## What is Crossing?

When Realm A calls a function that lives in Realm B, execution "crosses" from A
into B, and B becomes the realm that is acting. Inside that function, changes
are saved to B's storage. Once it returns, the acting realm goes back to A.

Gno follows this chain of calls, so at any point the code can ask:

- **Who started the transaction?** → ` + "`unsafe.OriginCaller()`" + `
- **Who called me?** → ` + "`unsafe.PreviousRealm()`" + `
- **Where am I?** → ` + "`unsafe.CurrentRealm()`" + `
- **Whose state is being modified?** → Always the current realm's state

To let other realms cross into a function, give it a ` + "`cur realm`" + ` parameter.
` + "`cur`" + ` is a handle to this realm's own context while the function runs:

` + "```go" + `
func ModifyState(cur realm, newValue string) {
    state = newValue  // Modifies THIS realm's state
}
` + "```" + `

The parameter alone does not cross. The caller crosses by marking the call with
` + "`cross`" + `, passing its own ` + "`cur`" + `:

` + "```go" + `
other.ModifyState(cross(cur), "hello")  // crosses into the "other" realm
` + "```" + `

Calling another realm's function without ` + "`cross`" + ` is rejected. The plain
` + "`fn(cur, ...)`" + ` form is only for calls within the same realm, where nothing crosses.

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
import (
    "testing"

    "gno.land/p/nt/testutils/v0"
)

func TestWithCaller(t *testing.T) {
    // Simulate PreviousRealm() being another realm
    testing.SetRealm(testing.NewCodeRealm("gno.land/r/caller"))

    // Simulate OriginCaller() being an end user
    testing.SetOriginCaller(testutils.TestAddress("user1"))
}
` + "```" + `

---

### See also:
- [Common Mistakes](/r/docs/mistakes) - OriginCaller vs cur.Previous()
`
}
