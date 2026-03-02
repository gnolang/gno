# CommonDAO Package Storage Extension

Storage package is an extension of `gno.land/p/nt/commondao/v0` that provides
alternative storage implementations.

## Member Storage

This custom implementation of `MemberStorage` is an implementation with
grouping support that automatically adds or removes members from the storage
when members are added or removed from any of the member groups.

The implementation provided by the `commondao` package doesn't automatically
add members to the storage when any of the groups change, if need then users
have to be added explicitly.

Adding or removing users automatically could be beneficial in some cases where
implementation requires iterating all unique users within the storage and
within each group, or counting all unique users within them. It also makes it
cheaper to iterate because having all users within the same storage doesn't
require to iterate each group.

Package also provide a `GetMemberGroups()` function that takes advantage of
this storage which can be used to return the names of the groups that an
account is a member of.

Example usage:

```go
import (
  "gno.land/p/nt/commondao/v0"
  "gno.land/p/nt/commondao/v0/exts/storage"
)

func main() {
  // Create a new member storage with grouping
  s := storage.NewMemberStorage()
  
  // Create a member group for moderators
  moderators, err := s.Grouping().Add("moderators")
  if err != nil {
    panic(err)
  }

  // Add members to the moderators group
  moderators.Members().Add("g1...a")
  moderators.Members().Add("g1...b")

  // Create a DAO that uses the new member storage
  dao := commondao.New(commondao.WithMemberStorage(s))

  // Output: 2
  println(dao.Members().Size())
}
```
