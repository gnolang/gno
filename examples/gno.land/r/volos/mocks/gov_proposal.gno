package main

import(
	"time"
	"gno.land/r/volos/gov/governance"
)

func main(){
	cb := func (){
		governance.SetProposalThreshold(100)
	}

	governance.CreateProposal(cross, "a", "a", time.Duration(1000), cb)
	
}