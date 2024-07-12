package govdao

import (
	"std"
)

// Status enum.
type Status string

var (
	Accepted  Status = "Accepted" // passed voting
	Active    Status = "Active"
	Rejected  Status = "Rejected"
	Expired   Status = "Expired"   // voting deadline passed
	Succeeded Status = "Succeeded" // has been executed
)

// Voter defines the needed methods for a voting system
type Voter interface {

	// IsAccepted indicates if the voting process had been accepted
	IsAccepted(voters []std.Address) bool

	// IsFinished indicates if the voting process is finished
	IsFinished(voters []std.Address) bool

	// Vote adds a new vote to the voting system
	Vote(voters []std.Address, caller std.Address, flag string)

	// Status returns a human friendly string describing how the voting process is going
	Status(voters []std.Address) string
}
