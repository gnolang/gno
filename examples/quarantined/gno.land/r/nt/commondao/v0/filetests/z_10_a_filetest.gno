// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_10_a_filetest

package z_10_a_filetest

import (
	"strconv"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var daoID uint64

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID = commondao.New(cross(cur), "*Pwn* [link](https://evil.example) # H1", "")
}

func main() {
	// DAO names are user controlled: markdown metacharacters must
	// render escaped, never as markup
	println(commondao.Render(strconv.FormatUint(daoID, 10)))
}

// Output:
// # \*Pwn\* \[link\]\(https://evil\.example\) \# H1
// [View Proposals](/r/nt/commondao/v0:2/proposals) • [View Settings](/r/nt/commondao/v0:2/settings)
//
// ---
// | Council |
// | --- |
// | g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 |
