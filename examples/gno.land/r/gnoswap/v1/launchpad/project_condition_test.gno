package launchpad

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestProjectCondition_Getters(t *testing.T) {
	tests := []struct {
		name             string
		tokenPath        string
		minimumAmount    int64
		getterFunc       string
		expectedValue    interface{}
		expectedHasError bool
		expectedError    string
	}{
		{
			name:          "get token path",
			tokenPath:     "gno.land/r/demo/gns",
			minimumAmount: 1000,
			getterFunc:    "TokenPath",
			expectedValue: "gno.land/r/demo/gns",
		},
		{
			name:          "get minimum amount",
			tokenPath:     "gno.land/r/demo/gns",
			minimumAmount: 1000,
			getterFunc:    "MinimumAmount",
			expectedValue: int64(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			condition := &ProjectCondition{
				tokenPath:     tt.tokenPath,
				minimumAmount: tt.minimumAmount,
			}

			// Execute and verify
			var result interface{}
			switch tt.getterFunc {
			case "TokenPath":
				result = condition.TokenPath()
			case "MinimumAmount":
				result = condition.MinimumAmount()
			}
			uassert.Equal(t, result, tt.expectedValue)
		})
	}
}

func TestProjectCondition_CheckCondition(t *testing.T) {
	tests := []struct {
		name             string
		tokenPath        string
		minimumAmount    int64
		inputToken       string
		inputAmount      int64
		expectedHasError bool
		expectedError    string
	}{
		{
			name:             "check condition succeed when input amount greater than minimum",
			tokenPath:        "gno.land/r/demo/gns",
			minimumAmount:    1000,
			inputToken:       "gno.land/r/demo/gns",
			inputAmount:      1500,
			expectedHasError: false,
		},
		{
			name:             "check condition succeed when input amount equals minimum",
			tokenPath:        "gno.land/r/demo/gns",
			minimumAmount:    1000,
			inputToken:       "gno.land/r/demo/gns",
			inputAmount:      1000,
			expectedHasError: false,
		},
		{
			name:             "check condition fail when input amount less than minimum",
			tokenPath:        "gno.land/r/demo/gns",
			minimumAmount:    1000,
			inputToken:       "gno.land/r/demo/gns",
			inputAmount:      500,
			expectedHasError: true,
			expectedError:    "input amount(500) is less than minimum amount(1000)",
		},
		{
			name:             "check condition fail when token path not matched",
			tokenPath:        "gno.land/r/demo/gns",
			minimumAmount:    1000,
			inputToken:       "gno.land/r/demo/btc",
			inputAmount:      1500,
			expectedHasError: true,
			expectedError:    "token path(gno.land/r/demo/btc) is not matched",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			condition := &ProjectCondition{
				tokenPath:     tt.tokenPath,
				minimumAmount: tt.minimumAmount,
			}

			// Execute
			err := condition.CheckCondition(tt.inputToken, tt.inputAmount)

			// Verify
			if tt.expectedHasError {
				uassert.ErrorContains(t, err, tt.expectedError)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestNewProjectCondition(t *testing.T) {
	tests := []struct {
		name              string
		tokenPath         string
		minimumAmount     int64
		expectedTokenPath string
		expectedMinAmount int64
		expectedHasError  bool
		expectedError     string
	}{
		{
			name:              "new project condition create valid condition with correct values",
			tokenPath:         "gno.land/r/demo/gns",
			minimumAmount:     1000,
			expectedTokenPath: "gno.land/r/demo/gns",
			expectedMinAmount: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			condition := NewProjectCondition(tt.tokenPath, tt.minimumAmount)

			// Verify
			uassert.Equal(t, condition.TokenPath(), tt.expectedTokenPath)
			uassert.Equal(t, condition.MinimumAmount(), tt.expectedMinAmount)
		})
	}
}
