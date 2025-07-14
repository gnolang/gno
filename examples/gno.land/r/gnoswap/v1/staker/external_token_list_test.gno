package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/rbac"
)

func TestAddToken(t *testing.T) {
	tests := []struct {
		name                  string
		caller                std.Address
		inputTokenPath        string
		expectedAllowedTokens []string
		expectedHasAbort      bool
		expectedAbortMsg      string
		expectedHasPanic      bool
		expectedPanicMsg      string
	}{
		{
			name:           "add token success by admin",
			caller:         testutils.TestAddress("admin"),
			inputTokenPath: "gno.land/r/gnoswap/v1/token1",
			expectedAllowedTokens: []string{
				consts.GNS_PATH,
				consts.GNOT,
				"gno.land/r/gnoswap/v1/token1",
			},
			expectedHasAbort: false,
			expectedAbortMsg: "",
			expectedHasPanic: false,
			expectedPanicMsg: "",
		},
		{
			name:           "add token failed by non admin",
			caller:         testutils.TestAddress("non-admin"),
			inputTokenPath: consts.GNS_PATH,
			expectedAllowedTokens: []string{
				consts.GNS_PATH,
				consts.GNOT,
			},
			expectedHasAbort: true,
			expectedAbortMsg: "caller(g1dehkuttpv3kkjmjlta047h6lta047h6lj8fjdj) has no permission",
			expectedHasPanic: false,
			expectedPanicMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			allowedTokens = defaultAllowed

			setupExternalTokenListAdminPermission(t)
			testing.SetOriginCaller(tt.caller)

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					AddToken(cross, tt.inputTokenPath)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					AddToken(cross, tt.inputTokenPath)
				})
			} else {
				AddToken(cross, tt.inputTokenPath)
			}

			// then
			uassert.Equal(t, len(tt.expectedAllowedTokens), len(allowedTokens))

			allowedTokenMap := make(map[string]bool)
			for _, token := range tt.expectedAllowedTokens {
				allowedTokenMap[token] = true
			}

			for _, token := range allowedTokens {
				uassert.True(t, allowedTokenMap[token])
			}
		})
	}
}

func TestRemoveToken(t *testing.T) {
	customTokenPath := "gno.land/r/gnoswap/v1/custom"

	currentDefaultTokens := make([]string, 0)
	currentDefaultTokens = append(currentDefaultTokens, defaultAllowed...)
	currentDefaultTokens = append(currentDefaultTokens, customTokenPath)

	tests := []struct {
		name                  string
		caller                std.Address
		inputTokenPath        string
		expectedAllowedTokens []string
		expectedHasAbort      bool
		expectedAbortMsg      string
		expectedHasPanic      bool
		expectedPanicMsg      string
	}{
		{
			name:                  "remove token success with custom token by admin",
			caller:                testutils.TestAddress("admin"),
			inputTokenPath:        customTokenPath,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT},
			expectedHasAbort:      false,
			expectedAbortMsg:      "",
			expectedHasPanic:      false,
			expectedPanicMsg:      "",
		},
		{
			name:                  "remove token failed by non admin",
			caller:                testutils.TestAddress("non-admin"),
			inputTokenPath:        customTokenPath,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT, customTokenPath},
			expectedHasAbort:      true,
			expectedAbortMsg:      "caller(g1dehkuttpv3kkjmjlta047h6lta047h6lj8fjdj) has no permission",
		},
		{
			name:                  "remove token failed by default token",
			caller:                testutils.TestAddress("admin"),
			inputTokenPath:        consts.GNS_PATH,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT, customTokenPath},
			expectedHasAbort:      true,
			expectedAbortMsg:      "[GNOSWAP-STAKER-018] can not delete default external token: can not remove default token(gno.land/r/gnoswap/v1/gns)",
			expectedHasPanic:      false,
			expectedPanicMsg:      "",
		},
		{
			// XXX: When removing a non-existent token, it is necessary to confirm that the operation was successful.
			name:                  "remove token failed by not existing token",
			caller:                testutils.TestAddress("admin"),
			inputTokenPath:        "not-existing-token",
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT, customTokenPath},
			expectedHasAbort:      false,
			expectedAbortMsg:      "",
			expectedHasPanic:      false,
			expectedPanicMsg:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowedTokens = currentDefaultTokens

			setupExternalTokenListAdminPermission(t)
			testing.SetOriginCaller(tt.caller)

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					RemoveToken(cross, tt.inputTokenPath)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					RemoveToken(cross, tt.inputTokenPath)
				})
			} else {
				RemoveToken(cross, tt.inputTokenPath)
			}

			// then
			uassert.Equal(t, len(tt.expectedAllowedTokens), len(allowedTokens))

			allowedTokenMap := make(map[string]bool)
			for _, token := range tt.expectedAllowedTokens {
				allowedTokenMap[token] = true
			}

			for _, token := range allowedTokens {
				uassert.True(t, allowedTokenMap[token])
			}
		})
	}
}

func TestModifyTokenList(t *testing.T) {
	tests := []struct {
		name                  string
		caller                std.Address
		inputTokenPath        string
		inputValidator        TokenValidator
		inputExecutor         TokenExecutor
		expectedAllowedTokens []string
		expectedHasError      bool
		expectedErrorMsg      string
	}{
		{
			name:                  "modify token list success with valid token addition",
			caller:                testutils.TestAddress("admin"),
			inputTokenPath:        "gno.land/r/gnoswap/v1/token1",
			inputValidator:        nil,
			inputExecutor:         addTokenExecutor,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT, "gno.land/r/gnoswap/v1/token1"},
			expectedHasError:      false,
			expectedErrorMsg:      "",
		},
		{
			name:           "modify token list success with validation failure",
			caller:         testutils.TestAddress("admin"),
			inputTokenPath: consts.GNS_PATH,
			inputValidator: func(tokenPath string) error {
				return ufmt.Errorf("validation failure")
			},
			inputExecutor:         addTokenExecutor,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT},
			expectedHasError:      true,
			expectedErrorMsg:      "validation failure",
		},
		{
			name:                  "modify token list failed by non admin",
			caller:                testutils.TestAddress("non-admin"),
			inputTokenPath:        "gno.land/r/gnoswap/v1/token1",
			inputValidator:        nil,
			inputExecutor:         addTokenExecutor,
			expectedAllowedTokens: []string{consts.GNS_PATH, consts.GNOT},
			expectedHasError:      true,
			expectedErrorMsg:      "caller(g1dehkuttpv3kkjmjlta047h6lta047h6lj8fjdj) has no permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowedTokens = defaultAllowed

			setupExternalTokenListAdminPermission(t)
			testing.SetOriginCaller(tt.caller)

			// when
			err := modifyTokenList(cross, tt.inputTokenPath, tt.inputValidator, tt.inputExecutor)
			if tt.expectedHasError {
				uassert.Error(t, err)
				uassert.Equal(t, tt.expectedErrorMsg, err.Error())
			} else {
				uassert.NoError(t, err)
			}

			// then
			uassert.Equal(t, len(tt.expectedAllowedTokens), len(allowedTokens))

			allowedTokenMap := make(map[string]bool)
			for _, token := range tt.expectedAllowedTokens {
				allowedTokenMap[token] = true
			}

			for _, token := range allowedTokens {
				uassert.True(t, allowedTokenMap[token])
			}
		})
	}
}

func setupExternalTokenListAdminPermission(t *testing.T) {
	t.Helper()

	addr := testutils.TestAddress("admin")
	setupExternalDepositFeePermission(t, addr, "admin")
}

func setupExternalTokenListGovernancePermission(t *testing.T) {
	t.Helper()

	addr := testutils.TestAddress("governance")
	setupExternalDepositFeePermission(t, addr, "governance")
}

func setupExternalTokenListPermission(t *testing.T, caller std.Address, permissionName string) {
	t.Helper()

	testing.SetOriginCaller(caller)
	testing.SetRealm(std.NewUserRealm(caller))

	globalManager := rbac.GlobalManager()
	globalManager.UpdatePermission(permissionName, "access", func(addr std.Address) error {
		if addr != caller {
			return ufmt.Errorf("caller is not %s", addr)
		}

		return nil
	})
}
