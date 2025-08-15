package main

import (
	"time"
	"gno.land/r/volos/gov/governance"
)

// callback function that will be executed if the proposal passes
func callback() {
	// This is a simple test callback that doesn't do anything
	// In a real scenario, this could update governance parameters, mint tokens, etc.
}

func main() {
	governance.CreateProposal(
		cross,
		"Test Proposal", 
		"This is a test proposal to verify governance functionality",
		time.Duration(5*time.Hour),
		callback,
	)

	governance.CreateProposal(
		cross,
		"Test Proposal", 
		"This is a test proposal to verify governance functionality",
		time.Duration(5*time.Hour),
		callback,
	)

	governance.CreateProposal(
		cross,
		"Test Proposal", 
		"This is a test proposal to verify governance functionality",
		time.Duration(5*time.Hour),
		callback,
	)
} 
