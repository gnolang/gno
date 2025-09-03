# p/moul/authz

TODO:
- transferable from static to dao-managed
- easy to plug
- easy to extend
- english-first usage
---
- in addition to PBAC -> RBAC
- try to have a shared interface between drivers

```go
// p/authz

// p/dao

// r/moul/config
var MyKeys authz.Membership = addrset.NewSet("g1manfred", "g1backup").Safe()
// XXX: how to manage MyKeys -> safe object
// r/team/dao

// r/team/config

// r/sys/team
func Has(handle string, addr std.Address)

type Membership struct {} // optional

// r/boards2

dao := commondao.New()
membersWithRoles := authz.WrapExistingAddrset(dao)
// membersWithRoles.AssignRole..
```

