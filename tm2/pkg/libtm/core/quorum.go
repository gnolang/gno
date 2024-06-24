package core

// hasSuperMajority verifies that there is a 2F+1 voting power majority
// in the given message set.
// This follows the constraint that N > 3F, i.e., the total voting power of faulty processes is smaller than
// one third of the total voting power
func (t *Tendermint) hasSuperMajority(messages []Message) bool {
	sumVotingPower := t.verifier.GetSumVotingPower(messages)
	totalVotingPower := t.verifier.GetTotalVotingPower(t.state.getHeight())

	return sumVotingPower > (2 * totalVotingPower / 3)
}

// hasFaultyMajority verifies that there is an F+1 voting power majority
// in the given message set.
// This follows the constraint that N > 3F, i.e., the total voting power of faulty processes is smaller than
// one third of the total voting power
func (t *Tendermint) hasFaultyMajority(messages []Message) bool {
	sumVotingPower := t.verifier.GetSumVotingPower(messages)
	totalVotingPower := t.verifier.GetTotalVotingPower(t.state.getHeight())

	return sumVotingPower > totalVotingPower/3
}
