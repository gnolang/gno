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

## ComposeDenom
Composes a qualified denomination string from the realm's pkg path and the provided base denomination. e.g `/gno.land/r/demo/blog:ugnot`. This method should be used when interacting with the `Banker` interface.

#### Parameters
- `denom` **string** - denomination to compose with the realm's pkg path

#### Usage
```go
denom := r.ComposeDenom("ugnot")
```

---

## ComposeRealmDenom
```go
func ComposeRealmDenom(pkgPath, denom string) string
```

Composes a qualified denomination string from the realm's pkg path and the provided base denomination. e.g `/gno.land/r/demo/blog:ugnot`. This method should be used when interacting with the `Banker` interface. It can also be used as a method of the `Realm` object, see [Realm.Denom](realm.md#composedenom).

#### Parameters
- `pkgPath` **string** - package path of the realm
- `denom` **string** - denomination to compose with the realm's pkg path

#### Usage
```go
denom := std.ComposeRealmDenom("gno.land/r/demo/blog", "ugnot")
```

