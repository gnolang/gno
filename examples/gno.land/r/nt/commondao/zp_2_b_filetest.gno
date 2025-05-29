package main

import (
	"gno.land/r/nt/commondao"
)

func main() {
	cross(commondao.CreateSubDAOProposal)(404, "", "")
}

// Error:
// DAO not found
