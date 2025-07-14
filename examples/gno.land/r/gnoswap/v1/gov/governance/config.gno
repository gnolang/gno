package governance

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"

	"gno.land/r/gnoswap/v1/halt"
)

// ReconfigureByAdmin updates the proposal realted configuration.
// Returns the new configuration version number.
func ReconfigureByAdmin(
	cur realm,
	votingStartDelay uint64,
	votingPeriod uint64,
	votingWeightSmoothingDuration uint64,
	quorum uint64,
	proposalCreationThreshold uint64,
	executionDelay uint64,
	executionWindow uint64,
) uint64 {
	caller := std.PreviousRealm().Address()
	if err := common.AdminOnly(caller); err != nil {
		panic(err)
	}

	return reconfigure(
		cur,
		votingStartDelay,
		votingPeriod,
		votingWeightSmoothingDuration,
		quorum,
		proposalCreationThreshold,
		executionDelay,
		executionWindow,
	)
}

// reconfigure updates the Governor's configuration.
// Only governance contract can execute this function via proposal
// Returns the new configuration version number.
func reconfigure(
	cur realm,
	votingStartDelay uint64,
	votingPeriod uint64,
	votingWeightSmoothingDuration uint64,
	quorum uint64,
	proposalCreationThreshold uint64,
	executionDelay uint64,
	executionWindow uint64,
) uint64 {
	if err := halt.IsHalted(); err != nil {
		panic(err)
	}

	en.MintAndDistributeGns(cross)

	// Update proposal states before reconfiguring
	transition := UpdateProposalStateTransition()
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	prevVersion := governanceState.configCounter

	newConfig := Config{
		VotingStartDelay:              votingStartDelay,
		VotingPeriod:                  votingPeriod,
		VotingWeightSmoothingDuration: votingWeightSmoothingDuration,
		Quorum:                        quorum,
		ProposalCreationThreshold:     proposalCreationThreshold,
		ExecutionDelay:                executionDelay,
		ExecutionWindow:               executionWindow,
	}

	// Apply reconfiguration using state transition
	transition = ReconfigureTransition(newConfig)
	if err := applyTransition(transition); err != nil {
		panic(err)
	}

	newVersion := governanceState.configCounter

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Reconfigure",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"votingStartDelay", formatUint(newConfig.VotingStartDelay),
		"votingPeriod", formatUint(newConfig.VotingPeriod),
		"votingWeightSmoothingDuration", formatUint(newConfig.VotingWeightSmoothingDuration),
		"quorum", formatUint(newConfig.Quorum),
		"proposalCreationThreshold", formatUint(newConfig.ProposalCreationThreshold),
		"executionDelay", formatUint(newConfig.ExecutionDelay),
		"executionPeriod", formatUint(newConfig.ExecutionWindow),
		"newConfigVersion", formatUint(newVersion),
		"prevConfigVersion", formatUint(prevVersion),
	)

	return newVersion
}

// GetConfigVersion returns the configuration for a specific version.
// If version is 0, it returns the current configuration.
func GetConfigVersion(version uint64) Config {
	cfg, exists := governanceState.GetConfig(version)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("config version(%d) does not exist", version),
		))
	}

	return cfg
}

// GetLatestConfig returns the latest configuration.
func GetLatestConfig() Config {
	return governanceState.currentConfig
}

// GetLatestConfigVersion returns the latest configuration version.
func GetLatestConfigVersion() uint64 {
	return governanceState.configCounter
}

// GetProposalCreationThreshold returns the current proposal creation threshold.
func GetProposalCreationThreshold() uint64 {
	return governanceState.currentConfig.ProposalCreationThreshold
}
