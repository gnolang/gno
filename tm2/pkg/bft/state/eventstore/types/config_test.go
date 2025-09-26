package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_GetParam(t *testing.T) {
	t.Parallel()

	const paramName = "param"

	testTable := []struct {
		name string
		cfg  *Config

		expectedParam any
	}{
		{
			"param not set",
			&Config{},
			nil,
		},
		{
			"valid param set",
			&Config{
				Params: map[string]any{
					paramName: 10,
				},
			},
			10,
		},
	}

	for _, testCase := range testTable {

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testCase.expectedParam, testCase.cfg.GetParam(paramName))
		})
	}
}
