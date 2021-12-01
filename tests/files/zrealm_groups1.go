// PKGPATH: gno.land/r/groups_test
package groups_test

import (
	"gno.land/p/groups"
	"gno.land/p/testutils"
)

var vset groups.VoteSet

func init() {
	addr1 := testutils.TestAddress("test1")
	addr2 := testutils.TestAddress("test2")
	vset = groups.NewVoteList()
	// XXX keep going...
	vset.SetVote(addr1, "yes")
	vset.SetVote(addr2, "yes")
}

func main() {
	println(vset.Size())
	println("yes:", vset.CountVotes("yes"))
	println("no:", vset.CountVotes("no"))
}

// Output:
// 2
// yes: 2
// no: 0
