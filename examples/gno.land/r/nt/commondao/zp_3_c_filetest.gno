package main

import (
	"gno.land/r/nt/commondao"
)

func main() {
	cross(commondao.CreateDismissalProposal)(404)
}

// Error:
// DAO not found
