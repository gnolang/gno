# realm_id: Realm and User Identification Utilities

`realmid` provides simple functions to identify callers in the Gno ecosystem - whether they're users or packages.

## Functions

```go
// Get the previous caller (user address or package path)
func Previous() string

// Get the current realm identifier  
func Current() string

// Check if an ID is a package path (contains dots)
func IsPackage(id string) bool

// Check if an ID is a user address (no dots)
func IsUser(id string) bool
```

## Basic Usage

```go
import "gno.land/p/samcrew/realmid"

func MyFunction() {
    caller := realmid.Previous()
    
    if realmid.IsUser(caller) {
        // caller is a user address like "g1abc123..."
        println("Called by user:", caller)
    } else if realmid.isPackage(caller) {
        // caller is a package path like "gno.land/p/demo/users"
        println("Called by package:", caller)
    } else {
        println("Should not happen")
    }
}
```

## Integration with basedao

```go
// Use realmid.Previous as the caller identifier function
config := &basedao.Config{
    Name:     "My DAO",
    CallerID: realmid.Previous,
    // ...
}
```