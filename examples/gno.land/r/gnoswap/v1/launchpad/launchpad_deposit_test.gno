package launchpad

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/gns"

	_ "gno.land/r/onbloc/obl"
)

func TestDepositGnsPrivate(t *testing.T) {
	userAddr := testutils.TestAddress("user")

	tests := []struct {
		name                 string
		projectID            string
		tierDuration         int64
		depositAmount        int64
		callerAddress        std.Address
		expectedHasError     bool
		expectedErrorMessage string
	}{
		{
			name:             "deposit gns successfully",
			projectID:        "gno.land/r/onbloc/obl:123",
			tierDuration:     30,
			depositAmount:    1000,
			callerAddress:    userAddr,
			expectedHasError: false,
		},
		{
			name:                 "deposit gns failed by invalid project id",
			projectID:            "invalid",
			tierDuration:         30,
			depositAmount:        1000,
			callerAddress:        userAddr,
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-003] requested data not found || project(invalid) not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initLaunchpadDepositTest(t)
			testing.SetOriginCaller(tt.callerAddress)

			_, _, _, _, err := depositGns(tt.projectID, tt.tierDuration, tt.depositAmount, tt.callerAddress)

			if tt.expectedHasError {
				uassert.Equal(t, tt.expectedErrorMessage, err.Error())
				return
			}

			uassert.NoError(t, err)
		})
	}
}

func TestStakeGovernance(t *testing.T) {
	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	launchpadAddr, _ := access.GetAddress(access.ROLE_LAUNCHPAD)

	tests := []struct {
		name                 string
		projectID            string
		depositAmount        int64
		callerAddress        std.Address
		userAddress          std.Address
		expectedError        bool
		expectedHasAbort     bool
		expectedAbortMessage string
	}{
		{
			name:                 "stake governance successfully",
			projectID:            "gno.land/r/onbloc/obl:123",
			depositAmount:        1000000,
			callerAddress:        launchpadAddr,
			userAddress:          adminAddr,
			expectedError:        false,
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
		{
			name:                 "stake governance failed by caller is not launchpad",
			projectID:            "gno.land/r/onbloc/obl:123",
			depositAmount:        1000000,
			callerAddress:        adminAddr,
			userAddress:          adminAddr,
			expectedError:        false,
			expectedHasAbort:     true,
			expectedAbortMessage: "caller(g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d) has no permission",
		},
		{
			name:                 "stake governance failed by invalid project id",
			projectID:            "invalid",
			depositAmount:        1000000,
			callerAddress:        launchpadAddr,
			userAddress:          adminAddr,
			expectedError:        true,
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
	}

	// caller is launchpad, user is admin
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initLaunchpadDepositTest(t)

			testing.SetOriginCaller(tt.userAddress)
			gns.Approve(cross, tt.callerAddress, tt.depositAmount)

			testing.SetOriginCaller(tt.callerAddress)

			var err error

			stakeGovernanceFn := func() {
				err = stakeGovernance(tt.projectID, tt.depositAmount, tt.callerAddress, tt.userAddress)
			}

			if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMessage, stakeGovernanceFn)
				return
			}

			stakeGovernanceFn()

			if tt.expectedError {
				uassert.Error(t, err)
				return
			}

			uassert.NoError(t, err)

		})
	}
}

func initLaunchpadDepositTest(t *testing.T) {
	testing.SetHeight(123)

	// Initialize test environment
	deposits = avl.NewTree()
	projects = avl.NewTree()
	projectTierRewardManagers = avl.NewTree()

	currentTime := time.Now().Unix()
	currentHeight := std.ChainHeight()

	// Setup test project
	project, _ := createProject(
		&createProjectParams{
			name:               "Test Project",
			tokenPath:          "gno.land/r/onbloc/obl",
			depositAmount:      1000000,
			tier30Ratio:        30,
			tier90Ratio:        30,
			tier180Ratio:       40,
			averageBlockTimeMs: 2000,
			recipient:          testutils.TestAddress("project"), // TODO: change to real address
			startTime:          currentTime + 10,
			currentTime:        currentTime,
			currentHeight:      currentHeight,
		},
	)
	projects.Set(project.ID(), project)

	testing.SkipHeights(10)
}
