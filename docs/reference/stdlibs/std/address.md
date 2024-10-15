---
id: address
---

# Address
Native address type in Gno, conforming to the Bech32 format.

```go
type Address string
func (a Address) String() string {...}
func (a Address) IsValid() bool {...}
```

## String
Get **string** representation of **Address**.

#### Usage
```go
stringAddr := addr.String()
```

---
## IsValid
Check if **Address** is of a valid format.

#### Usage
```go
if !address.IsValid() {...}
```
