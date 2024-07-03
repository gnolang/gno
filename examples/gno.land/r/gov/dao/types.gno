package govdao

import (
	"std"
)

// Status enum.
type Status string

var (
	Accepted    Status = "accepted"
	Active      Status = "active"
	NotAccepted Status = "not accepted"
	Expired     Status = "expired"
	Succeeded   Status = "succeeded"
)

// Voter defines the needed methods for a voting system
type Voter interface {
	Accepted(voters []std.Address) bool
	Finished(voters []std.Address) bool
	Vote(voters []std.Address, flag string)
}
