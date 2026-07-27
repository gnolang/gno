// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_15_a_filetest

package z_15_a_filetest

import (
	"chain"
	"strconv"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var (
	daoA uint64
	daoB uint64
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoA = commondao.New(cross(cur), "A", "")
	daoB = commondao.New(cross(cur), "B", "")
}

func main() {
	a := commondao.GetView(daoA).Address()
	b := commondao.GetView(daoB).Address()

	println("valid:", a.IsValid() && b.IsValid())
	println("distinct:", a != b)

	// The address is a pure derivation from the realm path and DAO ID
	derived := chain.DerivePkgSubAddr("gno.land/r/nt/commondao/v0", "dao/"+strconv.FormatUint(daoA, 10))
	println("derived:", a == derived)
}

// Output:
// valid: true
// distinct: true
// derived: true
