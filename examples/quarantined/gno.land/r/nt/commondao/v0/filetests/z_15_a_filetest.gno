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
	daoA = commondao.New(cross(cur), "A", "Purpose", "", "")
	daoB = commondao.New(cross(cur), "B", "Purpose", "", "")
}

func main() {
	a := commondao.GetView(daoA).Address()
	b := commondao.GetView(daoB).Address()

	println("valid:", a.IsValid() && b.IsValid())
	println("distinct:", a != b)

	// The address is a pure derivation from the realm path and DAO ID
	derived := chain.DerivePkgSubAddr("gno.land/r/nt/commondao/v0", "dao/"+strconv.FormatUint(daoA, 10))
	println("derived:", a == derived)

	// The genesis DAO carries a derived address too
	g := commondao.GetView(1).Address()
	println("genesis:", g == chain.DerivePkgSubAddr("gno.land/r/nt/commondao/v0", "dao/1"))
}

// Output:
// valid: true
// distinct: true
// derived: true
// genesis: true
