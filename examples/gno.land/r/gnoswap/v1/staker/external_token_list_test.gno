package staker

/*
func TestAddToken(t *testing.T) {
	adminAddr := consts.ADMIN
	nonAdminAddr := testutils.TestAddress("nonadmin")
	std.TestSetOrigCaller(adminAddr)

	tests := []struct {
		name        string
		caller      std.Address
		tokenPath   string
		expected    []string
		shouldPanic bool
	}{
		{
			name:        "Admin adds a new token",
			caller:      adminAddr,
			tokenPath:   "newTokenPath",
			expected:    append(defaultAllowed, "newTokenPath"),
			shouldPanic: false,
		},
		{
			name:        "Non-admin tries to add a token",
			caller:      nonAdminAddr,
			tokenPath:   "unauthorizedToken",
			expected:    defaultAllowed,
			shouldPanic: true,
		},
		{
			name:        "Admin adds an existing token",
			caller:      adminAddr,
			tokenPath:   consts.GNS_PATH,
			expected:    defaultAllowed,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetOrigCaller(tt.caller)

			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			AddToken(tt.tokenPath)

			if !tt.shouldPanic {
				for _, token := range tt.expected {
					if !contains(allowedTokens, token) {
						t.Errorf("Expected token %s not found in allowedTokens", token)
					}
				}
			}
		})
	}
}

func TestRemoveToken(t *testing.T) {
	adminAddr := consts.ADMIN
	nonAdminAddr := testutils.TestAddress("nonadmin")
	std.TestSetOrigCaller(adminAddr)

	tests := []struct {
		name        string
		caller      std.Address
		tokenPath   string
		expected    []string
		shouldPanic bool
	}{
		{
			name:        "Admin removes a custom token",
			caller:      adminAddr,
			tokenPath:   "customToken",
			expected:    defaultAllowed,
			shouldPanic: false,
		},
		{
			name:        "Non-admin tries to remove a token",
			caller:      nonAdminAddr,
			tokenPath:   "unauthorizedToken",
			expected:    allowedTokens,
			shouldPanic: true,
		},
		{
			name:        "Admin tries to remove a default token",
			caller:      adminAddr,
			tokenPath:   consts.GNOT,
			expected:    allowedTokens,
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetOrigCaller(tt.caller)

			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				}
			}()

			RemoveToken(tt.tokenPath)

			if !tt.shouldPanic {
				for _, token := range tt.expected {
					if !contains(allowedTokens, token) {
						t.Errorf("Expected token %s not found in allowedTokens", token)
					}
				}
			}
		})
	}
}
*/
