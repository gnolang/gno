package launchpad

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/obl"

	_ "gno.land/r/onbloc/bar"
	_ "gno.land/r/onbloc/foo"
)

func TestCreateProject(t *testing.T) {
	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	launchpadAddr, _ := access.GetAddress(access.ROLE_LAUNCHPAD)
	oblPath := "gno.land/r/onbloc/obl"
	tests := []struct {
		name                 string
		projectName          string
		tokenPath            string
		recipient            std.Address
		depositAmount        int64
		conditionTokens      string
		conditionAmounts     string
		tier30Ratio          int64
		tier90Ratio          int64
		tier180Ratio         int64
		startTime            int64
		callerAddress        std.Address
		averageBlockTimeMs   int64
		expectedProjectID    string
		expectedStartHeight  int64
		expectedHasAbort     bool
		expectedAbortMessage string
	}{
		{
			name:                 "create project successfully",
			projectName:          "Obl Protocol",
			tokenPath:            oblPath,
			recipient:            testutils.TestAddress("projectAddr"),
			depositAmount:        1_000_000_000,
			conditionTokens:      "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionAmounts:     "1000*PAD*2000",
			tier30Ratio:          10,
			tier90Ratio:          20,
			tier180Ratio:         70,
			startTime:            time.Now().Unix() + 3600, // 1 hour
			callerAddress:        adminAddr,
			averageBlockTimeMs:   2000,
			expectedProjectID:    "gno.land/r/onbloc/obl:123",
			expectedStartHeight:  std.ChainHeight() + 3600/2, // 1 hour / 2
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
		{
			name:                 "create project successfully by average block time",
			projectName:          "Obl Protocol",
			tokenPath:            oblPath,
			recipient:            testutils.TestAddress("projectAddr"),
			depositAmount:        1_000_000_000,
			conditionTokens:      "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionAmounts:     "1000*PAD*2000",
			tier30Ratio:          10,
			tier90Ratio:          20,
			tier180Ratio:         70,
			startTime:            time.Now().Unix() + 3600, // 1 hour
			callerAddress:        adminAddr,
			averageBlockTimeMs:   1000,
			expectedProjectID:    "gno.land/r/onbloc/obl:123",
			expectedStartHeight:  std.ChainHeight() + 3600,
			expectedHasAbort:     false,
			expectedAbortMessage: "",
		},
		{
			name:                 "create project failed by not admin",
			projectName:          "Obl Protocol",
			tokenPath:            oblPath,
			recipient:            testutils.TestAddress("projectAddr"),
			depositAmount:        1_000_000_000,
			conditionTokens:      "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionAmounts:     "1000*PAD*2000",
			tier30Ratio:          10,
			tier90Ratio:          20,
			tier180Ratio:         70,
			startTime:            time.Now().Unix() + 3600, // 1 hour
			callerAddress:        testutils.TestAddress("non-admin"),
			averageBlockTimeMs:   2000,
			expectedHasAbort:     true,
			expectedAbortMessage: "[GNOSWAP-LAUNCHPAD-002] invalid address || caller(g1dehkuttpv3kkjmjlta047h6lta047h6lj8fjdj) is not authorized for role admin",
		},
		{
			name:                 "create project failed by start time is in the past",
			projectName:          "Obl Protocol",
			tokenPath:            oblPath,
			recipient:            testutils.TestAddress("projectAddr"),
			depositAmount:        1_000_000_000,
			conditionTokens:      "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionAmounts:     "1000*PAD*2000",
			tier30Ratio:          10,
			tier90Ratio:          20,
			tier180Ratio:         70,
			startTime:            time.Now().Unix() - 3600, // 1 hour
			callerAddress:        adminAddr,
			averageBlockTimeMs:   2000,
			expectedHasAbort:     true,
			expectedAbortMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || start time(1234564290) must be greater than now(1234571490)",
		},
		{
			name:                 "create project failed by start time is less than minimum start delay time",
			projectName:          "Obl Protocol",
			tokenPath:            oblPath,
			recipient:            testutils.TestAddress("projectAddr"),
			depositAmount:        1_000_000_000,
			conditionTokens:      "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionAmounts:     "1000*PAD*2000",
			tier30Ratio:          10,
			tier90Ratio:          20,
			tier180Ratio:         70,
			startTime:            time.Now().Unix() + 3600 - 1, // 1 hour
			callerAddress:        adminAddr,
			averageBlockTimeMs:   2000,
			expectedHasAbort:     true,
			expectedAbortMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || start time(1234571489) must be greater than now(1234571490)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initLaunchpadProjectTest(t)
			initLaunchpadProjectTestAvgBlockTime(t, tt.averageBlockTimeMs)
			testing.SetOriginCaller(tt.callerAddress)

			obl.Approve(cross, launchpadAddr, int64(tt.depositAmount))

			createProjectFn := func() {
				CreateProject(
					cross,
					tt.projectName,
					tt.tokenPath,
					tt.recipient,
					tt.depositAmount,
					tt.conditionTokens,
					tt.conditionAmounts,
					tt.tier30Ratio,
					tt.tier90Ratio,
					tt.tier180Ratio,
					tt.startTime,
				)
			}

			if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMessage, createProjectFn)
				return
			}

			createProjectFn()

			project, err := getProject(tt.expectedProjectID)
			uassert.NoError(t, err)

			uassert.Equal(t, project.ID(), tt.expectedProjectID)
			uassert.Equal(t, project.StartHeight(), tt.expectedStartHeight)
		})
	}
}

func TestCreateProjectPrivate(t *testing.T) {
	launchpadAddr, _ := access.GetAddress(access.ROLE_LAUNCHPAD)
	projectAddr := testutils.TestAddress("projectAddr")
	oblPath := "gno.land/r/onbloc/obl"
	tests := []struct {
		name                 string
		params               *createProjectParams
		expectedProjectID    string
		expectedHasError     bool
		expectedErrorMessage string
	}{
		{
			name: "create project successfully",
			params: &createProjectParams{
				name:               "Obl Protocol",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(70),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600), // 1 hour
				averageBlockTimeMs: 2000,
			},
			expectedProjectID:    "gno.land/r/onbloc/obl:123",
			expectedHasError:     false,
			expectedErrorMessage: "",
		},
		{
			name: "create project successfully by 30:40:30 ratio",
			params: &createProjectParams{
				name:               "Obl Protocol",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(30),
				tier90Ratio:        int64(40),
				tier180Ratio:       int64(30),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600),
				averageBlockTimeMs: 2000,
			},
			expectedProjectID:    "gno.land/r/onbloc/obl:123",
			expectedHasError:     false,
			expectedErrorMessage: "",
		},
		{
			name: "create project failed by invalid ratio sum (less than 100%)",
			params: &createProjectParams{
				name:               "Invalid Ratio",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(60),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600),
				averageBlockTimeMs: 2000,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || invalid ratio, sum of all tiers(30:10, 90:20, 180:60) should be 100",
		},
		{
			name: "create project failed by project name is empty",
			params: &createProjectParams{
				name:               "",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(70),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600),
				averageBlockTimeMs: 2000,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || project name cannot be empty",
		},
		{
			name: "create project failed by project owner address is invalid",
			params: &createProjectParams{
				name:               "TTT",
				tokenPath:          oblPath,
				recipient:          std.Address("abcdef"),
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(70),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600),
				averageBlockTimeMs: 2000,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-002] invalid address || recipient address(abcdef)",
		},
		{
			name: "create project failed by invalid start time",
			params: &createProjectParams{
				name:                  "Obl Protocol",
				tokenPath:             oblPath,
				recipient:             projectAddr,
				depositAmount:         1_000_000_000,
				conditionTokens:       "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:      "1000*PAD*2000",
				tier30Ratio:           int64(10),
				tier90Ratio:           int64(20),
				tier180Ratio:          int64(70),
				currentTime:           int64(time.Now().Unix()),
				currentHeight:         std.ChainHeight(),
				startTime:             int64(time.Now().Unix() + 3600 - 1), // 1 hour - 1 second
				averageBlockTimeMs:    2000,
				minimumStartDelayTime: 3600,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || start time(1234571489) must be greater than now(1234571490)",
		},
		{
			name: "create project failed by invalid token path",
			params: &createProjectParams{
				name:                  "Obl Protocol",
				tokenPath:             "gno.land/r/test/token",
				recipient:             projectAddr,
				depositAmount:         1_000_000_000,
				conditionTokens:       "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:      "1000*PAD*2000",
				tier30Ratio:           int64(10),
				tier90Ratio:           int64(20),
				tier180Ratio:          int64(70),
				currentTime:           int64(time.Now().Unix()),
				currentHeight:         std.ChainHeight(),
				startTime:             int64(time.Now().Unix() + 3600 - 1), // 1 hour - 1 second
				averageBlockTimeMs:    2000,
				minimumStartDelayTime: 3600,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-007] invalid input data || tokenPath(gno.land/r/test/token) not registered",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initLaunchpadProjectTest(t)
			testing.SetOriginCaller(projectAddr)

			obl.Approve(cross, launchpadAddr, int64(tt.params.depositAmount))

			projectId, err := createProject(tt.params)

			if tt.expectedHasError {
				uassert.Equal(t, err.Error(), tt.expectedErrorMessage)
				return
			}

			uassert.Equal(t, projectId.ID(), tt.expectedProjectID)
		})
	}
}

func TestCreateProjectPrivate_Multiple(t *testing.T) {
	launchpadAddr, _ := access.GetAddress(access.ROLE_LAUNCHPAD)
	projectAddr := testutils.TestAddress("projectAddr")
	oblPath := "gno.land/r/onbloc/obl"
	tests := []struct {
		name                 string
		setupProjectIds      []string
		params               *createProjectParams
		expectedProjectID    string
		expectedHasError     bool
		expectedErrorMessage string
	}{
		{
			name: "create project successfully",
			setupProjectIds: []string{
				"gno.land/r/onbloc/obl:1",
			},
			params: &createProjectParams{
				name:               "Obl Protocol",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(70),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600), // 1 hour
				averageBlockTimeMs: 2000,
			},
			expectedProjectID:    "gno.land/r/onbloc/obl:123",
			expectedHasError:     false,
			expectedErrorMessage: "",
		},
		{
			name: "create project failed by duplicate project id",
			setupProjectIds: []string{
				"gno.land/r/onbloc/obl:123",
			},
			params: &createProjectParams{
				name:               "Obl Protocol",
				tokenPath:          oblPath,
				recipient:          projectAddr,
				depositAmount:      1_000_000_000,
				conditionTokens:    "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
				conditionAmounts:   "1000*PAD*2000",
				tier30Ratio:        int64(10),
				tier90Ratio:        int64(20),
				tier180Ratio:       int64(70),
				currentTime:        int64(time.Now().Unix()),
				currentHeight:      std.ChainHeight(),
				startTime:          int64(time.Now().Unix() + 3600), // 1 hour
				averageBlockTimeMs: 2000,
			},
			expectedHasError:     true,
			expectedErrorMessage: "[GNOSWAP-LAUNCHPAD-008] can not create same project in same block || project(gno.land/r/onbloc/obl:123) already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initLaunchpadProjectTest(t)
			for _, projectID := range tt.setupProjectIds {
				setupLaunchpadProjectTestEmptyProject(t, projectID)
			}

			obl.Approve(cross, launchpadAddr, int64(tt.params.depositAmount))

			projectId, err := createProject(tt.params)

			if tt.expectedHasError {
				uassert.Equal(t, err.Error(), tt.expectedErrorMessage)
				return
			}

			uassert.Equal(t, projectId.ID(), tt.expectedProjectID)
		})
	}
}

func TestCreateProjectParams_ValidatName(t *testing.T) {
	tests := []struct {
		name          string
		projectName   string
		expectedError error
	}{
		{
			name:          "project name is valid",
			projectName:   "Obl Protocol",
			expectedError: nil,
		},
		{
			name:          "project name is empty",
			projectName:   "",
			expectedError: errInvalidInput,
		},
		{
			name:          "project name is too long",
			projectName:   strings.Repeat("a", 101), // 101 characters
			expectedError: errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				name: test.projectName,
			}

			err := params.validateName()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateTokenPath(t *testing.T) {
	tests := []struct {
		name          string
		tokenPath     string
		expectedError error
	}{
		{
			name:          "token path is valid",
			tokenPath:     "gno.land/r/onbloc/obl",
			expectedError: nil,
		},
		{
			name:          "token path is empty",
			tokenPath:     "",
			expectedError: errInvalidInput,
		},
		{
			name:          "token path is not registered",
			tokenPath:     "gno.land/r/test/non-registered",
			expectedError: errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				tokenPath: test.tokenPath,
			}

			err := params.validateTokenPath()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateRecipient(t *testing.T) {
	tests := []struct {
		name          string
		recipient     std.Address
		expectedError error
	}{
		{
			name:          "recipient is valid",
			recipient:     testutils.TestAddress("projectAddr"),
			expectedError: nil,
		},
		{
			name:          "recipient is invalid",
			recipient:     std.Address("invalid"),
			expectedError: errInvalidAddress,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				recipient: test.recipient,
			}

			err := params.validateRecipient()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateDepositAmount(t *testing.T) {
	tests := []struct {
		name          string
		depositAmount int64
		expectedError error
	}{
		{
			name:          "deposit amount is valid",
			depositAmount: 1_000_000_000,
			expectedError: nil,
		},
		{
			name:          "deposit amount is zero",
			depositAmount: 0,
			expectedError: errInvalidInput,
		},
		{
			name:          "deposit amount is negative",
			depositAmount: -1,
			expectedError: errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				depositAmount: test.depositAmount,
			}

			err := params.validateDepositAmount()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateRatio(t *testing.T) {
	tests := []struct {
		name          string
		tier30Ratio   int64
		tier90Ratio   int64
		tier180Ratio  int64
		expectedError error
	}{
		{
			name:          "ratio is valid by sum is 100",
			tier30Ratio:   10,
			tier90Ratio:   20,
			tier180Ratio:  70,
			expectedError: nil,
		},
		{
			name:          "ratio is invalid by sum is not 100",
			tier30Ratio:   10,
			tier90Ratio:   20,
			tier180Ratio:  60,
			expectedError: errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				tier30Ratio:  test.tier30Ratio,
				tier90Ratio:  test.tier90Ratio,
				tier180Ratio: test.tier180Ratio,
			}

			err := params.validateRatio()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateStartTime(t *testing.T) {
	projectMinimumStartDelayTime := int64(60 * 60) // 1 hour

	tests := []struct {
		name          string
		startTime     int64
		expectedError error
	}{
		{
			name:          "start time is valid",
			startTime:     int64(time.Now().Unix() + projectMinimumStartDelayTime),
			expectedError: nil,
		},
		{
			name:          "start time is invalid by in the past",
			startTime:     int64(time.Now().Unix() - 10),
			expectedError: errInvalidInput,
		},
		{
			name:          "start time is invalid by equal to now",
			startTime:     int64(time.Now().Unix()),
			expectedError: errInvalidInput,
		},
		{
			name:          "start time is invalid by less than minimum start delay time",
			startTime:     int64(time.Now().Unix() + projectMinimumStartDelayTime - 1),
			expectedError: errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				startTime: test.startTime,
			}

			err := params.validateStartTime(int64(time.Now().Unix()), projectMinimumStartDelayTime)

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

func TestCreateProjectParams_ValidateConditions(t *testing.T) {
	tests := []struct {
		name             string
		conditionTokens  string
		conditionAmounts string
		expectedError    error
	}{
		{
			name:             "conditions are valid",
			conditionTokens:  "gno.land/r/onbloc/obl*PAD*gno.land/r/onbloc/bar",
			conditionAmounts: "1000*PAD*2000",
			expectedError:    nil,
		},
		{
			name:             "conditions are invalid by not registered token path",
			conditionTokens:  "gno.land/r/onbloc/obl*PAD*gno.land/r/test/token",
			conditionAmounts: "1000*PAD*2000",
			expectedError:    errInvalidInput,
		},
		{
			name:             "conditions are invalid by zero amount",
			conditionTokens:  "gno.land/r/onbloc/obl",
			conditionAmounts: "0",
			expectedError:    errInvalidInput,
		},
		{
			name:             "conditions are invalid by condition count mismatch",
			conditionTokens:  "gno.land/r/onbloc/obl*PAD*gno.land/r/onbloc/bar",
			conditionAmounts: "1000*PAD*2000*PAD*3000",
			expectedError:    errInvalidInput,
		},
		{
			name:             "conditions are invalid by duplicated token path",
			conditionTokens:  "gno.land/r/onbloc/obl*PAD*gno.land/r/onbloc/bar*PAD*gno.land/r/onbloc/obl",
			conditionAmounts: "1000*PAD*2000*PAD*3000",
			expectedError:    errInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			params := &createProjectParams{
				conditionTokens:  test.conditionTokens,
				conditionAmounts: test.conditionAmounts,
			}

			err := params.validateConditions()

			if test.expectedError == nil {
				uassert.NoError(t, err)
			} else {
				uassert.ErrorContains(t, err, test.expectedError.Error())
			}
		})
	}
}

// define test methods for launchpad project
func initLaunchpadProjectTest(t *testing.T) {
	testing.SetHeight(123)
	projects = new(avl.Tree)
	projectTierRewardManagers = new(avl.Tree)
}

func initLaunchpadProjectTestAvgBlockTime(t *testing.T, avgBlockTimeMs int64) {
	emissionAddr, _ := access.GetAddress(access.ROLE_EMISSION)
	testing.SetOriginCaller(emissionAddr)
	gns.SetCallbackEmissionChange(cross, func(amount int64) {})

	adminAddr, _ := access.GetAddress(access.ROLE_ADMIN)
	testing.SetOriginCaller(adminAddr)
	gns.SetAvgBlockTimeInMsByAdmin(cross, avgBlockTimeMs)
}

func setupLaunchpadProjectTestEmptyProject(t *testing.T, projectID string) {
	project := &Project{id: projectID}
	projects.Set(projectID, project)
}
