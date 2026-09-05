// PKGPATH: gno.land/r/test/rendererr
package rendererr

// Render is reachable unauthenticated via vm/qrender with an attacker-chosen
// path. Two pre-existing defects on that path are covered here:
//   - the "invalid proposal id" error echoed the raw pid segment (strconv
//     quotes it but leaves markdown/HTML metachars), so a crafted pid injected
//     inline markup into a govDAO-branded page;
//   - a control byte in the query string made the pager's MustGetPageByPath
//     panic, aborting the landing-page render.
// Both are now escaped / degraded-gracefully.

import (
	"strings"
	"testing"

	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/impl"
	"gno.land/r/gov/dao/v3/memberstore"
)

const user address = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func init(cur realm) {
	memberstore.Get(0, cur).DeleteAll()
	memberstore.Get(0, cur).SetTier(memberstore.T1)
	memberstore.Get(0, cur).SetMember(memberstore.T1, user, memberstore.NewMember(3))
	dao.UpdateImpl(cross(cur), dao.NewUpdateRequest(impl.NewGovDAO(), nil))
}

func main(cur realm) {
	testing.SetOriginCaller(user)
	testing.SetRealm(testing.NewUserRealm(user))

	// 1. A non-numeric pid whose segment carries an inline-markdown payload.
	// No '/', so it routes to {pid}; ParseInt fails and the error is escaped.
	pout := dao.Render(cross(cur), "99[pwn](evil.example)")
	println("payload is escaped, not a live link:",
		strings.Contains(pout, `\[pwn\]`) && !strings.Contains(pout, "[pwn](evil.example)"))
	println("still shows the invalid-id message:",
		strings.Contains(pout, "Invalid proposal ID format"))

	// Same on the /votes route.
	vout := dao.Render(cross(cur), "88[x](evil.example)/votes")
	println("votes route escapes it too:",
		!strings.Contains(vout, "[x](evil.example)"))

	// 2. A control byte in the query — url.Parse rejects it. The landing page
	// must degrade to page 1, not panic.
	lout := dao.Render(cross(cur), "?\x01")
	println("control-byte query renders instead of panicking:",
		strings.Contains(lout, "# GovDAO"))
}

// Output:
// payload is escaped, not a live link: true
// still shows the invalid-id message: true
// votes route escapes it too: true
// control-byte query renders instead of panicking: true
