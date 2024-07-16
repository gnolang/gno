---
id: realm
---

# Realm
Structure representing a realm in Gno. See concept page [here](../../../concepts/realms.md). 

```go
type Realm struct {
    addr    Address
    pkgPath string
}

func (r Realm) Addr() Address {...}
func (r Realm) PkgPath() string {...}
func (r Realm) IsUser() bool {...}
```

## Addr
Returns the **Address** field of the realm it was called upon.

#### Usage
```go
realmAddr := r.Addr() // eg. g1n2j0gdyv45aem9p0qsfk5d2gqjupv5z536na3d
```
---
## PkgPath
Returns the **string** package path of the realm it was called upon.

#### Usage
```go
realmPath := r.PkgPath() // eg. gno.land/r/gnoland/blog
```
---
## IsUser
Checks if the realm it was called upon is a user realm.

#### Usage
```go
if r.IsUser() {...}
```
