// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_23_d_filetest

// The home index with nothing listed.
//
// pager.New treats a total of zero as having no valid page — including
// page 1, which its own Picker links to — so the home route skips the
// pager entirely when the listed set is empty. Without that, unlisting
// every DAO turns "?page=1" into an error page.
package z_23_d_filetest

import (
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx, genesis council
)

func init(cur realm) {
	// The genesis DAO lists itself at creation; unlist it so nothing is
	// listed. Only its own council member may toggle the flag.
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.SetListed(cross(cur), 1, false)
}

func main(cur realm) {
	println("listed:", commondao.IsListed(1))

	// Both render the header alone — no list, and no "invalid page number".
	println(commondao.Render(""))
	println(commondao.Render("?page=1"))
}

// Output:
// listed: false
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
//
//
// # Common DAO
// ---
// This realm can be used to create CommonDAO instances based on [commondao](/p/nt/commondao/v0/) package.
