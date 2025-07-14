package governance

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/rbac"
)

var (
	adminAddr    = testutils.TestAddress("admin")
	notAdminAddr = testutils.TestAddress("notadmin")
)

func TestInitialConfig(t *testing.T) {
	tests := []struct {
		name           string
		config         Config
		expectedConfig Config
	}{
		{
			name:   "initial config is set",
			config: createMockConfig(),
			expectedConfig: Config{
				VotingStartDelay:              100000,
				VotingPeriod:                  700000,
				VotingWeightSmoothingDuration: 90000,
				Quorum:                        60,
				ProposalCreationThreshold:     2_000_000_000,
				ExecutionDelay:                90000,
				ExecutionWindow:               3000000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			// Save original state and create test state
			originalState := governanceState
			testState := NewGovernanceState(tt.config)

			// Temporarily replace global state
			governanceState = testState

			// Restore original state after test
			defer func() {
				governanceState = originalState
			}()

			// when
			currentConfig := governanceState.currentConfig

			// then
			uassert.Equal(t, currentConfig.VotingStartDelay, tt.expectedConfig.VotingStartDelay)
			uassert.Equal(t, currentConfig.VotingPeriod, tt.expectedConfig.VotingPeriod)
			uassert.Equal(t, currentConfig.VotingWeightSmoothingDuration, tt.expectedConfig.VotingWeightSmoothingDuration)
			uassert.Equal(t, currentConfig.Quorum, tt.expectedConfig.Quorum)
			uassert.Equal(t, currentConfig.ProposalCreationThreshold, tt.expectedConfig.ProposalCreationThreshold)
			uassert.Equal(t, currentConfig.ExecutionDelay, tt.expectedConfig.ExecutionDelay)
			uassert.Equal(t, currentConfig.ExecutionWindow, tt.expectedConfig.ExecutionWindow)
		})
	}
}

func TestReconfigureByAdmin(t *testing.T) {
	tests := []struct {
		name             string
		caller           std.Address
		config           Config
		newConfig        Config
		expectedHasPanic bool
		expectedHasAbort bool
		expectedPanicMsg string
		expectedAbortMsg string
	}{
		{
			name:   "reconfigure by admin is success",
			caller: adminAddr,
			config: createMockConfig(),
			newConfig: Config{
				VotingStartDelay:              0,
				VotingPeriod:                  0,
				VotingWeightSmoothingDuration: 0,
				Quorum:                        0,
				ProposalCreationThreshold:     0,
				ExecutionDelay:                0,
				ExecutionWindow:               0,
			},
			expectedHasPanic: false,
			expectedHasAbort: false,
		},
		{
			name:   "reconfigure by admin is failed by non-admin",
			caller: notAdminAddr,
			config: createMockConfig(),
			newConfig: Config{
				VotingStartDelay:              0,
				VotingPeriod:                  0,
				VotingWeightSmoothingDuration: 0,
				Quorum:                        0,
				ProposalCreationThreshold:     0,
				ExecutionDelay:                0,
				ExecutionWindow:               0,
			},
			expectedHasPanic: false,
			expectedHasAbort: true,
			expectedAbortMsg: "caller(g1dehhgctyd45kuh6lta047h6lta047h6lsn8gun) has no permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			// Save original state and create test state
			originalState := governanceState
			testState := NewGovernanceState(tt.config)

			// Temporarily replace global state
			governanceState = testState

			// Restore original state after test
			defer func() {
				governanceState = originalState
			}()

			setupConfigAdminPermission(t)

			// when
			reconfigureByAdminFn := func() {
				testing.SetOriginCaller(tt.caller)
				testing.SetRealm(std.NewUserRealm(tt.caller))

				ReconfigureByAdmin(
					cross,
					tt.newConfig.VotingStartDelay,
					tt.newConfig.VotingPeriod,
					tt.newConfig.VotingWeightSmoothingDuration,
					tt.newConfig.Quorum,
					tt.newConfig.ProposalCreationThreshold,
					tt.newConfig.ExecutionDelay,
					tt.newConfig.ExecutionWindow,
				)
			}

			// then
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, reconfigureByAdminFn)
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, reconfigureByAdminFn)
			} else {
				uassert.NotPanics(t, reconfigureByAdminFn)

				uassert.Equal(t, uint64(0), tt.newConfig.VotingStartDelay)
				uassert.Equal(t, uint64(0), tt.newConfig.VotingPeriod)
				uassert.Equal(t, uint64(0), tt.newConfig.VotingWeightSmoothingDuration)
				uassert.Equal(t, uint64(0), tt.newConfig.Quorum)
				uassert.Equal(t, uint64(0), tt.newConfig.ProposalCreationThreshold)
				uassert.Equal(t, uint64(0), tt.newConfig.ExecutionDelay)
				uassert.Equal(t, uint64(0), tt.newConfig.ExecutionWindow)
			}
		})
	}
}

func TestGetConfigVersion(t *testing.T) {
	tests := []struct {
		name                                  string
		inputVersion                          uint64
		expectedVotingStartDelay              uint64
		expectedVotingPeriod                  uint64
		expectedVotingWeightSmoothingDuration uint64
		expectedQuorum                        uint64
		expectedProposalCreationThreshold     uint64
		expectedExecutionDelay                uint64
		expectedExecutionWindow               uint64
		expectedHasPanic                      bool
		expectedPanicMsg                      string
	}{
		{
			name:                                  "get config version is success by current config (input version: 0)",
			inputVersion:                          0,
			expectedVotingStartDelay:              1,
			expectedVotingPeriod:                  1,
			expectedVotingWeightSmoothingDuration: 1,
			expectedQuorum:                        1,
			expectedProposalCreationThreshold:     1,
			expectedExecutionDelay:                1,
			expectedExecutionWindow:               1,
			expectedHasPanic:                      false,
		},
		{
			name:                                  "get config version is success by existing version",
			inputVersion:                          1,
			expectedVotingStartDelay:              0,
			expectedVotingPeriod:                  0,
			expectedVotingWeightSmoothingDuration: 0,
			expectedQuorum:                        0,
			expectedProposalCreationThreshold:     0,
			expectedExecutionDelay:                0,
			expectedExecutionWindow:               0,
			expectedHasPanic:                      false,
		},
		{
			name:                                  "get config version is failed by non-existent version",
			inputVersion:                          999,
			expectedVotingStartDelay:              0,
			expectedVotingPeriod:                  0,
			expectedVotingWeightSmoothingDuration: 0,
			expectedQuorum:                        0,
			expectedProposalCreationThreshold:     0,
			expectedExecutionDelay:                0,
			expectedExecutionWindow:               0,
			expectedHasPanic:                      true,
			expectedPanicMsg:                      "[GNOSWAP-GOVERNANCE-003] requested data not found || config version(999) does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			// Save original state and create test state
			originalState := governanceState

			// Create initial config
			config1 := Config{
				VotingStartDelay:              0,
				VotingPeriod:                  0,
				VotingWeightSmoothingDuration: 0,
				Quorum:                        0,
				ProposalCreationThreshold:     0,
				ExecutionDelay:                0,
				ExecutionWindow:               0,
			}

			// Create test state with initial config
			testState := NewGovernanceState(config1)

			// Add second config version
			config2 := Config{
				VotingStartDelay:              1,
				VotingPeriod:                  1,
				VotingWeightSmoothingDuration: 1,
				Quorum:                        1,
				ProposalCreationThreshold:     1,
				ExecutionDelay:                1,
				ExecutionWindow:               1,
			}
			testState = testState.WithNewConfig(config2)

			// Temporarily replace global state
			governanceState = testState

			// Restore original state after test
			defer func() {
				governanceState = originalState
			}()

			// when
			var result Config

			getConfigVersionFn := func() {
				result = GetConfigVersion(tt.inputVersion)
			}

			// then
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, getConfigVersionFn)
			} else {
				uassert.NotPanics(t, getConfigVersionFn)

				uassert.Equal(t, tt.expectedVotingStartDelay, result.VotingStartDelay)
				uassert.Equal(t, tt.expectedVotingPeriod, result.VotingPeriod)
				uassert.Equal(t, tt.expectedVotingWeightSmoothingDuration, result.VotingWeightSmoothingDuration)
				uassert.Equal(t, tt.expectedQuorum, result.Quorum)
				uassert.Equal(t, tt.expectedProposalCreationThreshold, result.ProposalCreationThreshold)
				uassert.Equal(t, tt.expectedExecutionDelay, result.ExecutionDelay)
				uassert.Equal(t, tt.expectedExecutionWindow, result.ExecutionWindow)
			}
		})
	}
}

func setupConfigAdminPermission(t *testing.T) {
	t.Helper()

	adminAddr := testutils.TestAddress("admin")

	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(std.NewUserRealm(adminAddr))

	globalManager := rbac.GlobalManager()
	globalManager.UpdatePermission("admin", "access", func(addr std.Address) error {
		if addr != adminAddr {
			return ufmt.Errorf("caller is not admin")
		}

		return nil
	})
}

func createMockConfig() Config {
	return Config{
		VotingStartDelay:              100000,
		VotingPeriod:                  700000,
		VotingWeightSmoothingDuration: 90000,
		Quorum:                        60,
		ProposalCreationThreshold:     2_000_000_000,
		ExecutionDelay:                90000,
		ExecutionWindow:               3000000,
	}
}
