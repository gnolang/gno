package nameregistry_test

import (
	"testing"

	"gno.land/p/nt/uassert"

	"gno.land/r/sys/nameregistry"
)

func TestIsNameValid(t *testing.T) {
	tests := []struct {
		name         string
		registerName string
		errMsg       string
	}{
		{
			name:         "valid",
			registerName: "test123",
		},
		{
			name:         "invalid prefix g1",
			registerName: "g1test",
			errMsg:       "name must not start these prefixes: g1, gi, gl",
		},
		{
			name:         "invalid prefix gi",
			registerName: "gitest",
			errMsg:       "name must not start these prefixes: g1, gi, gl",
		},
		{
			name:         "invalid prefix gl",
			registerName: "gltest",
			errMsg:       "name must not start these prefixes: g1, gi, gl",
		},
		{
			name:         "upercase not allowed",
			registerName: "TEST123",
			errMsg:       "name characters must be in lowercase",
		},
		{
			name:         "name too long",
			registerName: "alongtestname123",
			errMsg:       "name is too long, max length is 14",
		},
		{
			name:         "name too short",
			registerName: "a12",
			errMsg:       "name is too short, min length is 4",
		},
		{
			name:         "only numbers",
			registerName: "123456",
			errMsg:       `name must start with a letter between "a" and "z"`,
		},
		{
			name:         "only letters",
			registerName: "test",
			errMsg:       "name must end with three numbers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := nameregistry.IsNameValid(tt.registerName)

			// Assert
			if tt.errMsg != "" {
				uassert.ErrorContains(t, err, tt.errMsg, "expect an error")
			} else {
				uassert.Nil(t, err, "expect no error")
			}
		})
	}
}
