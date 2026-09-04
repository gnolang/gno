// PKGPATH: gno.land/r/nt/commondao/v0/filetests/z_23_a_filetest

// The proposals list route: pagination, and the view/sort toggles.
//
// The load-bearing assertion is on page 2: the "View:" and "Sort by:"
// links must NOT carry the current page. Switching to a view with fewer
// pages while page=2 is pinned would land on "invalid page number".
package z_23_a_filetest

import (
	"strconv"
	"testing"

	"gno.land/r/nt/commondao/v0"
)

const (
	owner = address("g16jpf0puufcpcjkph5nxueec8etpcldz7zwgydq") // @devx
	user  = address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5") // @test1
)

var listPath string

func init(cur realm) {
	testing.SetRealm(testing.NewUserRealm(owner))
	commondao.Invite(cross(cur), user)

	testing.SetRealm(testing.NewUserRealm(user))
	daoID := commondao.New(cross(cur), "Foo", "Purpose", "", "")
	listPath = strconv.FormatUint(daoID, 10) + "/proposals"

	// Nine proposals over a page size of eight: two pages.
	for i := 0; i < 9; i++ {
		commondao.CreateTextProposal(cross(cur), daoID, "Proposal "+strconv.Itoa(i+1), "Body", 7)
	}
}

func main() {
	// Page 1, then page 2 — where the toggle links must drop the page.
	println(commondao.Render(listPath))
	println(commondao.Render(listPath + "?page=2"))

	// Oldest-first, and the (empty) finished view.
	println(commondao.Render(listPath + "?order=asc"))
	println(commondao.Render(listPath + "?finished="))
}

// Output:
// # Foo: Proposals
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// View: [finished](/r/nt/commondao/v0:2/proposals?finished=) • Sort by: [oldest](/r/nt/commondao/v0:2/proposals?order=asc)
//
// **[#9 Proposal 9](/r/nt/commondao/v0:2/proposals/9)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=9&reason=&vote=)
//
// **[#8 Proposal 8](/r/nt/commondao/v0:2/proposals/8)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=8&reason=&vote=)
//
// **[#7 Proposal 7](/r/nt/commondao/v0:2/proposals/7)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=7&reason=&vote=)
//
// **[#6 Proposal 6](/r/nt/commondao/v0:2/proposals/6)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=6&reason=&vote=)
//
// **[#5 Proposal 5](/r/nt/commondao/v0:2/proposals/5)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=5&reason=&vote=)
//
// **[#4 Proposal 4](/r/nt/commondao/v0:2/proposals/4)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=4&reason=&vote=)
//
// **[#3 Proposal 3](/r/nt/commondao/v0:2/proposals/3)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=3&reason=&vote=)
//
// **[#2 Proposal 2](/r/nt/commondao/v0:2/proposals/2)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=2&reason=&vote=)
//
// ---
// \- | page 1 of 2 | [»](?page=2)
// # Foo: Proposals
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// View: [finished](/r/nt/commondao/v0:2/proposals?finished=) • Sort by: [oldest](/r/nt/commondao/v0:2/proposals?order=asc)
//
// **[#1 Proposal 1](/r/nt/commondao/v0:2/proposals/1)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=1&reason=&vote=)
//
// ---
// [«](?page=1) | page 2 of 2 | \-
// # Foo: Proposals
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// View: [finished](/r/nt/commondao/v0:2/proposals?finished=&order=asc) • Sort by: [newest](/r/nt/commondao/v0:2/proposals?order=desc)
//
// **[#1 Proposal 1](/r/nt/commondao/v0:2/proposals/1)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=1&reason=&vote=)
//
// **[#2 Proposal 2](/r/nt/commondao/v0:2/proposals/2)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=2&reason=&vote=)
//
// **[#3 Proposal 3](/r/nt/commondao/v0:2/proposals/3)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=3&reason=&vote=)
//
// **[#4 Proposal 4](/r/nt/commondao/v0:2/proposals/4)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=4&reason=&vote=)
//
// **[#5 Proposal 5](/r/nt/commondao/v0:2/proposals/5)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=5&reason=&vote=)
//
// **[#6 Proposal 6](/r/nt/commondao/v0:2/proposals/6)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=6&reason=&vote=)
//
// **[#7 Proposal 7](/r/nt/commondao/v0:2/proposals/7)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=7&reason=&vote=)
//
// **[#8 Proposal 8](/r/nt/commondao/v0:2/proposals/8)**
// Created by g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
// Voting ends on Fri, 20 Feb 2009 11:31pm UTC
// Votes: **0** • Status: **active** • [Vote](/r/nt/commondao/v0$help&func=Vote&daoID=2&proposalID=8&reason=&vote=)
//
// ---
// \- | page 1 of 2 | [»](?order=asc&page=2)
// # Foo: Proposals
// [Go to DAO](/r/nt/commondao/v0:2)
//
// ---
// View: [active](/r/nt/commondao/v0:2/proposals) • Sort by: [oldest](/r/nt/commondao/v0:2/proposals?finished=&order=asc)
//
// Currently there are no finished proposals
