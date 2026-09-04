// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_23_b_filetest

// The home index paginates over the LISTED set, not over every DAO.
//
// Listing is opt-in and off by default, so the newest DAOs are usually
// unlisted. Paginating over all DAOs and filtering the window afterwards
// drops listed DAOs off page 1 entirely (and, when a whole window is
// unlisted, renders an empty page with no pager at all — stranding them
// on pages the reader can no longer reach).
package z_23_b_filetest

import (
	"strconv"
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

	// Twenty DAOs (IDs 2..21). List the ten oldest; the rest stay
	// unlisted, so a window taken over all DAOs would start on entries
	// that never render.
	//
	// The counts are deliberately far apart: 11 listed is 2 pages, 21
	// total is 3. Were the pager sized on daos.Size() the extra page
	// would exist and render empty, so ?page=3 below discriminates the
	// pagination total — not just the iteration source.
	for i := 0; i < 20; i++ {
		id := commondao.New(cross(cur), "DAO "+strconv.Itoa(i+1), "Purpose", "", "")
		if i < 10 {
			commondao.SetListed(cross(cur), id, true)
		}
	}
}

func main() {
	// Eleven listed DAOs (ten above plus the genesis DAO) over a page
	// size of ten: page 1 is full and page 2 holds the remainder.
	println(commondao.Render(""))
	println(commondao.Render("?page=2"))

	// Past the end of the LISTED set: there is no third page. Sized on
	// every DAO there would be, and it would render empty.
	println(commondao.Render("?page=3"))
}

// Output:
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
//
// Here is a list of some of the DAOs that were created:
//
// - [DAO 10](/r/nt/commondao/v0:11)
// - [DAO 9](/r/nt/commondao/v0:10)
// - [DAO 8](/r/nt/commondao/v0:9)
// - [DAO 7](/r/nt/commondao/v0:8)
// - [DAO 6](/r/nt/commondao/v0:7)
// - [DAO 5](/r/nt/commondao/v0:6)
// - [DAO 4](/r/nt/commondao/v0:5)
// - [DAO 3](/r/nt/commondao/v0:4)
// - [DAO 2](/r/nt/commondao/v0:3)
// - [DAO 1](/r/nt/commondao/v0:2)
//
//
// \- | page 1 of 2 | [»](?page=2)
//
//
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
//
// Here is a list of some of the DAOs that were created:
//
// - [Common DAO](/r/nt/commondao/v0:1)
//
//
// [«](?page=1) | page 2 of 2 | \-
//
//
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
//
// invalid page number
