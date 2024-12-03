package users

import (
	"gno.land/p/demo/ufmt"
	susers "gno.land/p/sys/users"
)

// User extends the p/sys/users struct for more custom data
type User struct {
	base         *susers.User
	Bio          string
	Ordinal      int   // user's number
	RegisteredAt int64 // block height at time of registration
}

func (u *User) Render() string {
	out := ufmt.Sprintf("# User #%d - %s\n\n", u.Ordinal, u.base.Name)
	out += ufmt.Sprintf("## Bio:\n\n%s\n\n", u.Bio)
	out += ufmt.Sprintf("## Stats:\n\n")
	out += ufmt.Sprintf("- Address: %s\n\n", u.base.Address)
	out += ufmt.Sprintf("- Registered at block #%d\n\n", u.RegisteredAt)

	return out
}
