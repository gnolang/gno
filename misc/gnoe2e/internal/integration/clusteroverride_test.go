package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A setting named on the command line wins over the one a script declares, and
// the caller owns the consequence: asking for two validators runs a scenario
// that declared three, and that scenario may well go red.
func TestClusterOverridesApply(t *testing.T) {
	declared := ClusterSpec{
		Validators:           3,
		CodeSubmissionPolicy: "inert",
		BlockMaxGas:          20_000_000,
	}

	tests := []struct {
		name      string
		overrides ClusterOverrides
		want      ClusterSpec
	}{
		{
			name: "nothing named leaves the declaration alone",
			want: declared,
		},
		{
			name:      "a smaller validator count wins, red scenario or not",
			overrides: ClusterOverrides{Validators: ptr(2)},
			want: ClusterSpec{
				Validators:           2,
				CodeSubmissionPolicy: "inert",
				BlockMaxGas:          20_000_000,
			},
		},
		{
			name: "an approver named as an address replaces one named as a role",
			overrides: ClusterOverrides{
				PkgApprover: ptr("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			},
			want: ClusterSpec{
				Validators:           3,
				CodeSubmissionPolicy: "inert",
				PkgApprover:          "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
				BlockMaxGas:          20_000_000,
			},
		},
		{
			name: "every setting at once",
			overrides: ClusterOverrides{
				Validators:           ptr(1),
				CodeSubmissionPolicy: ptr("permissionless"),
				BlockMaxGas:          ptr(int64(3_000_000_000)),
			},
			want: ClusterSpec{
				Validators:           1,
				CodeSubmissionPolicy: "permissionless",
				BlockMaxGas:          3_000_000_000,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.overrides.Apply(declared))
			assert.Equal(t, ClusterSpec{
				Validators:           3,
				CodeSubmissionPolicy: "inert",
				BlockMaxGas:          20_000_000,
			}, declared, "the declaration itself is not modified")
		})
	}
}

func ptr[T any](v T) *T { return &v }
