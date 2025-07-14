package governance

// Global state instance
var governanceState *GovernanceState

func init() {
	// Initialize with default config
	initialConfig := Config{
		VotingStartDelay:              uint64(86400),         // 1d
		VotingPeriod:                  uint64(604800),        // 7d
		VotingWeightSmoothingDuration: uint64(86400),         // 1d
		Quorum:                        uint64(50),            // 50% of total xGNS supply
		ProposalCreationThreshold:     uint64(1_000_000_000), // 1_000_000_000
		ExecutionDelay:                uint64(86400),         // 1d
		ExecutionWindow:               uint64(2592000),       // 30d
	}

	governanceState = NewGovernanceState(initialConfig)
}

// applyTransition applies a state transition and updates the global state
func applyTransition(transition Transition) error {
	newState, err := transition(governanceState)
	if err != nil {
		return err
	}
	governanceState = newState
	return nil
}
