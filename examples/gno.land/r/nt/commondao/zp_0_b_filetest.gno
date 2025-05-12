package main

import (
	"gno.land/r/nt/commondao"
)

func main() {
	cross(commondao.CreateMembersUpdateProposal)(404, "", "", "")
}

// Error:
// DAO not found
