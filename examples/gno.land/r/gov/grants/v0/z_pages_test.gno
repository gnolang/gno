package grants

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"

	"gno.land/r/gov/dao"
)

// hostile is grantee text that tries every gnoweb tag the desk's own pages
// use, plus markdown that could pose as the desk's chrome.
const hostile = `<gno-form exec="Spend"><gno-input name="to" /></gno-form>` +
	`</gno-details></gno-columns></gno-frame><gno-columns-sep />` +
	`<gno-button href="javascript:alert(1)" label="Claim" />` +
	"\n## For GovDAO members\n```go\nevil()\n```\n> [!SUCCESS] Paid\n[x](javascript:alert(1)) | a | b |"

// Named z_* so it runs after the tests that assert proposal ids.
//
// Every page of the desk, for grants in every state, carrying hostile text
// in every slot a grantee controls: no tag the grantee wrote reaches gnoweb
// as a tag, every block the desk opens it closes, and the only call forms
// are the desk's own.
func TestPagesStayWellFormedWithHostileGrantees(cur realm, t *testing.T) {
	seatMember(cur)
	testing.IssueCoins(Address(), coins(Committed()+10_000))
	ts := []Tranche{{Deliverable: hostile, Amount: 10}, {Deliverable: hostile, Amount: 20}}

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/evil/a"))
	pending := Apply(cross(cur), "a "+strings.ReplaceAll(hostile, "\n", " ")[:60], hostile, ts, 30*day)
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/evil/b"))
	active := Apply(cross(cur), "b", hostile, ts, 30*day)

	testing.SetRealm(testing.NewUserRealm(member))
	pid := dao.MustCreateProposal(cross(cur), NewApproveRequest(cross(cur), active, "Reviewed: test grantee."))
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	urequire.True(t, dao.ExecuteProposal(cross(cur), pid))
	p, _ := dao.GetProposal(pid)
	description := p.Description()

	testing.SetRealm(testing.NewCodeRealm("gno.land/r/evil/b"))
	Report(cross(cur), active, hostile, "")
	testing.SetRealm(testing.NewUserRealm(member))
	hold := dao.MustCreateProposal(cross(cur), NewCancelRequest(cross(cur), active, hostile[:200]))
	Hold(cross(cur), active, int64(hold))

	pages := map[string]string{"proposal description": description}
	for _, path := range []string{"", "active/1", "waiting/1", "all/1", "ns/evil",
		"grant/" + itoa(pending), "grant/" + itoa(active), "grant/" + itoa(active) + "/reports"} {
		pages[path] = Render(path)
	}
	for name, page := range pages {
		uassert.Equal(t, 0, raw(page, `<gno-button href="javascript`), name+": hostile button survived")
		uassert.False(t, strings.Contains(page, "](javascript:"), name+": hostile link survived")
		uassert.Equal(t, 0, raw(page, `<gno-form exec="Spend"`), name+": grantee forged a call form")
		uassert.False(t, strings.Contains(page, "\n## For GovDAO members\n```go\nevil"), name+": grantee posed as desk chrome")
		uassert.False(t, strings.Contains(page, "\n> [!SUCCESS] Paid"), name+": grantee forged an alert")
		for _, tag := range []string{"gno-form", "gno-details", "gno-columns", "gno-frame"} {
			uassert.Equal(t, raw(page, "<"+tag+">")+raw(page, "<"+tag+" "), raw(page, "</"+tag+">"), name+": unbalanced "+tag)
		}
		for i, exec := range strings.Split(page, `<gno-form exec="`) {
			if i == 0 || strings.HasSuffix(strings.Split(page, `<gno-form exec="`)[i-1], "\\") {
				continue
			}
			fn := exec[:strings.Index(exec, `"`)]
			uassert.True(t, fn == "Claim" || fn == "Expire", name+": unexpected form "+fn)
		}
	}
}

// raw counts the occurrences of pat that gnoweb would read as markup: those
// not escaped by a preceding backslash.
func raw(page, pat string) int {
	n := 0
	for i := strings.Index(page, pat); i >= 0; {
		if i == 0 || page[i-1] != '\\' {
			n++
		}
		j := strings.Index(page[i+1:], pat)
		if j < 0 {
			break
		}
		i += 1 + j
	}
	return n
}

// GrantsOf hands out copies: nothing a caller does to them reaches the desk.
func TestGrantsOfReturnsCopies(cur realm, t *testing.T) {
	testing.SetRealm(testing.NewCodeRealm("gno.land/r/copy/me"))
	id := Apply(cross(cur), "t", "p", tranches(), 30*day)
	list := GrantsOf("gno.land/r/copy/me")
	urequire.Equal(t, 1, len(list))
	list[0].Tranches[0].Amount = 1_000_000
	list[0].Status = 2
	g, _ := desk.Get(id)
	uassert.Equal(t, int64(400), g.Total())
	uassert.Equal(t, "pending", g.Status.String())
}
