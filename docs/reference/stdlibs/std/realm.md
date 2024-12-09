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
func (r Realm) ComposeDenom(denom string) string {...}
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

## Denom
Composes a qualified denomination string from the realm's pkg path and the provided base denomination. e.g `/gno.land/r/demo/blog:ugnot`. This method should be used when interacting with the `Banker` interface.

#### Parameters
- `denom` **string** - The base denomination used to build the qualified denomination. Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
denom := r.Denom("ugnot")
```

---

## RealmDenom
```go
func RealmDenom(pkgPath, denom string) string
```

Composes a qualified denomination string from the realm's pkg path and the provided base denomination. e.g `/gno.land/r/demo/blog:ugnot`. This method should be used when interacting with the `Banker` interface. It can also be used as a method of the `Realm` object, see [Realm.Denom](realm.md#denom).

#### Parameters
- `pkgPath` **string** - package path of the realm
- `denom` **string** - The base denomination used to build the qualified denomination.  Must start with a lowercase letter, followed by 2–15 lowercase letters or digits.

#### Usage
```go
denom := std.RealmDenom("gno.land/r/demo/blog", "ugnot")
```

