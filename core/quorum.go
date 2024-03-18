package core

// hasSuperMajority verifies that there is a 2F+1 voting power majority
// in the given message set.
// This follows the constraint that N > 3f, i.e., the total voting power of faulty processes is smaller than
// one third of the total voting power
func (t *Tendermint) hasSuperMajority(messages []Message) bool {
	return t.verifier.GetSumVotingPower(messages) > (2 * t.verifier.GetTotalVotingPower(t.state.getHeight()) / 3)
}

// hasFaultyMajority verifies that there is an F+1 voting power majority
// in the given message set.
// This follows the constraint that N > 3f, i.e., the total voting power of faulty processes is smaller than
// one third of the total voting power
func (t *Tendermint) hasFaultyMajority(messages []Message) bool {
	return t.verifier.GetSumVotingPower(messages) > (t.verifier.GetTotalVotingPower(t.state.getHeight()) / 3)
}
