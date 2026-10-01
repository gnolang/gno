package main

import (
	"go/scanner"
	"go/token"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseGoBuildErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		output          string
		expectedError   error
		expectedErrorIs error
	}{
		{
			name:          "empty output",
			output:        "",
			expectedError: nil,
		},
		{
			name:   "random output",
			output: "xxx",
			expectedError: scanner.ErrorList{
				&scanner.Error{
					Msg: "Additional go build errors:\nxxx",
				},
			},
		},
		{
			name: "some errors",
			output: `xxx
main.gno:6:2: nasty error
pkg/file.gno:60:20: ugly error`,
			expectedError: scanner.ErrorList{
				&scanner.Error{
					Pos: token.Position{
						Filename: "main.gno",
						Line:     6,
						Column:   2,
					},
					Msg: "nasty error",
				},
				&scanner.Error{
					Pos: token.Position{
						Filename: "pkg/file.gno",
						Line:     60,
						Column:   20,
					},
					Msg: "ugly error",
				},
				&scanner.Error{
					Msg: "Additional go build errors:\nxxx",
				},
			},
		},
		{
			name:            "line parse error",
			output:          `main.gno:9000000000000000000000000000000000000000000000000000:11: error`,
			expectedErrorIs: strconv.ErrRange,
		},
		{
			name:            "column parse error",
			output:          `main.gno:1:9000000000000000000000000000000000000000000000000000: error`,
			expectedErrorIs: strconv.ErrRange,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := parseGoBuildErrors(tt.output)
			if eis := tt.expectedErrorIs; eis != nil {
				assert.ErrorIs(t, err, eis)
			} else {
				assert.Equal(t, tt.expectedError, err)
			}
		})
	}
}
