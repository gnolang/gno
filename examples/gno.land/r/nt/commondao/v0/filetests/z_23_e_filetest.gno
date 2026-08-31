// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_23_e_filetest

// A pager error raised inside a helper must be its own markdown block.
//
// renderCouncil is called mid-page by renderDAO, so its early return
// resumes the caller, which goes straight on to write the Treasury
// heading. Written as a bare string the error has no trailing newline
// and the two glue together ("invalid page number## Treasury"),
// destroying the H2 and taking the council table with it. This is the
// one pager site where that is observable — everywhere else the error
// is the last thing on the page, where the missing newline is trimmed.
package z_23_e_filetest

import (
	"strings"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	commondao.New(cross(cur), "Foo", "Purpose", "", "")
}

func main(cur realm) {
	// A council page far past the end.
	out := commondao.Render("2?members=99")

	// Slice around the boundary: the error and the heading that follows
	// must be separate blocks.
	i := strings.Index(out, "invalid page number")
	println(out[i : i+len("invalid page number\n\n## Treasury")])
}

// Output:
// invalid page number
//
// ## Treasury
