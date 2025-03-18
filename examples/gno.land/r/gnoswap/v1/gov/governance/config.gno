package governance

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/common"
	en "gno.land/r/gnoswap/v1/emission"

	"gno.land/r/gnoswap/v1/halt"
)

var (
	config         Config
	configVersions = avl.NewTree() // version -> Config
)

func init() {
	// https://docs.gnoswap.io/core-concepts/governance
	config = Config{
		VotingStartDelay:              uint64(86400),         // 1d
		VotingPeriod:                  uint64(604800),        // 7d
		VotingWeightSmoothingDuration: uint64(86400),         // 1d
		Quorum:                        uint64(50),            // 50% of total xGNS supply
		ProposalCreationThreshold:     uint64(1_000_000_000), // 1_000_000_000
		ExecutionDelay:                uint64(86400),         // 1d
		ExecutionWindow:               uint64(2592000),       // 30d
	}

	// config version 0 should return the current config
	// therefore we set initial config version to 1
	setConfigVersion(1, config)
}

func setConfigVersion(v uint64, cfg Config) {
	configVersions.Set(formatUint(v), cfg)
}

func getConfigByVersion(v uint64) (Config, bool) {
	value, exists := configVersions.Get(formatUint(v))
	if !exists {
		return Config{}, false
	}
	return value.(Config), true
}

func getLatestVersion() uint64 {
	var maxVersion uint64
	configVersions.ReverseIterate("", "", func(key string, value interface{}) bool {
		maxVersion = parseUint64(key)
		return true
	})
	return maxVersion
}

// ReconfigureByAdmin updates the proposal realted configuration.
// Returns the new configuration version number.
func ReconfigureByAdmin(
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

	en.MintAndDistributeGns()
	updateProposalsState()

	prevVersion := getLatestVersion()
	newVersion := prevVersion + 1

	config = Config{
		VotingStartDelay:              votingStartDelay,
		VotingPeriod:                  votingPeriod,
		VotingWeightSmoothingDuration: votingWeightSmoothingDuration,
		Quorum:                        quorum,
		ProposalCreationThreshold:     proposalCreationThreshold,
		ExecutionDelay:                executionDelay,
		ExecutionWindow:               executionWindow,
	}
	setConfigVersion(newVersion, config)

	prevAddr, prevPkgPath := getPrev()
	std.Emit(
		"Reconfigure",
		"prevAddr", prevAddr,
		"prevRealm", prevPkgPath,
		"votingStartDelay", formatUint(config.VotingStartDelay),
		"votingPeriod", formatUint(config.VotingPeriod),
		"votingWeightSmoothingDuration", formatUint(config.VotingWeightSmoothingDuration),
		"quorum", formatUint(config.Quorum),
		"proposalCreationThreshold", formatUint(config.ProposalCreationThreshold),
		"executionDelay", formatUint(config.ExecutionDelay),
		"executionPeriod", formatUint(config.ExecutionWindow),
		"newConfigVersion", formatUint(newVersion),
		"prevConfigVersion", formatUint(prevVersion),
	)

	return newVersion
}

// GetConfigVersion returns the configuration for a specific version.
// If version is 0, it returns the current configuration.
func GetConfigVersion(version uint64) Config {
	if version == 0 {
		return config
	}

	cfg, exists := getConfigByVersion(version)
	if !exists {
		panic(addDetailToError(
			errDataNotFound,
			ufmt.Sprintf("config version(%d) does not exist", version),
		))
	}

	return cfg
}

// GetLatestConfig() returns the latest configuration.
func GetLatestConfig() Config {
	return config
}

// GetLatestConfigVersion() returns the latest configuration version.
func GetLatestConfigVersion() uint64 {
	return getLatestVersion()
}

// GetProposalCreationThreshold() returns the current proposal creation threshold.
func GetProposalCreationThreshold() uint64 {
	return config.ProposalCreationThreshold
}
