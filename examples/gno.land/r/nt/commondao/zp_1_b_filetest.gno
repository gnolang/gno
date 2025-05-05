package main

import (
	"gno.land/r/nt/commondao"
)

func main() {
	cross(commondao.CreateTextProposal)(404, "", "", 0)
}

// Error:
// DAO not found
