// PKGPATH: gno.land/p/archives/groups
package groups

import (
	"gno.land/p/archives/groups"
	"gno.land/p/nt/testutils"
)

var vset groups.VoteSet

func init() {
	addr1 := testutils.TestAddress("test1")
	addr2 := testutils.TestAddress("test2")
	vset = groups.NewVoteList()
	vset.SetVote(addr1, "yes")
	vset.SetVote(addr2, "yes")
}

func main() {
	println(vset.Size())
	println("yes:", vset.CountVotes("yes"))
	println("no:", vset.CountVotes("no"))
}

// Error:
// gno.land/p/archives/groups/z_1.gno:9:10-24: name VoteSet not declared
