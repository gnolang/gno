package consensus

func (_ newRoundStepInfo) AssertWALMessage() {} // state.go
func (_ msgInfo) AssertWALMessage()          {} // state.go
func (_ timeoutInfo) AssertWALMessage()      {} // state.go
