package main

import (
	"gno.land/r/aib/ibc/core"
)

type X struct{}

func (X) ClientType() string {
	return "wrong--"
}

func (X) ValidateBasic() error {
	return nil
}

// CreateClient w/ invalid client type
func main() {
	core.CreateClient(cross, X{}, X{})
}

// Error:
// client type does not meet naming constraints: not a valid clientID: wrong---0
