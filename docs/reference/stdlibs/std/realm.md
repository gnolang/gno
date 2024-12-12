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
func (r Realm) CoinDenom(coinName string) string {...}
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

---

## CoinDenom
Composes a qualified denomination string from the realm's `pkgPath` and the provided coin name, e.g. `/gno.land/r/demo/blog:ugnot`. This method should be used to get fully qualified denominations of coins when interacting with the `Banker` module.

#### Parameters
- `coinName` **string** - The coin name used to build the qualified denomination. Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
denom := r.CoinDenom("ugnot")
```

