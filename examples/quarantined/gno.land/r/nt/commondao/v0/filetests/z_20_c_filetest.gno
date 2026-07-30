// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_20_c_filetest

package z_20_c_filetest

import (
	"strconv"
	"strings"
	"testing"

	pdao "gno.land/p/nt/commondao/v0"

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
	daoID = commondao.New(cross(cur), "Foo", "Purpose", "", "")

	pID := commondao.CreateRegisterKindProposal(cross(cur), daoID, "execution")
	commondao.Vote(cross(cur), daoID, pID, pdao.ChoiceYes, "")
	commondao.Execute(cross(cur), daoID, pID)
}

func main(cur realm) {
	testing.SetRealm(testing.NewUserRealm(user))

	// The execution kind's title and body are raw, user-supplied text
	// (executionDef, a package type that cannot implement the realm's
	// trustedMarkdownBody marker). The renderer must escape both so
	// markdown metacharacters render as text, never as markup.
	pID := commondao.CreateExecutionProposal(
		cross(cur),
		daoID,
		"*Pwn* # H1",
		"[link](https://evil.example) # Injected",
		func(_ int, _ realm) error { return nil },
	)

	out := commondao.Render(strconv.FormatUint(daoID, 10) + "/proposals/" + strconv.FormatUint(pID, 10))

	// Escaped forms are present; raw markup is not.
	println("title escaped:", strings.Contains(out, `\*Pwn\* \# H1`))
	println("body escaped:", strings.Contains(out, `\[link\]\(https://evil\.example\) \# Injected`))
	println("raw H1 injected:", strings.Contains(out, "\n# Injected"))
}

// Output:
// title escaped: true
// body escaped: true
// raw H1 injected: false
