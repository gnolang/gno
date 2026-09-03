package integration

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func script(clusterSection string) []byte {
	return []byte("# a scenario\ngnokey query auth/accounts\n\n" +
		"-- cluster --\n" + clusterSection)
}

func TestParseClusterSpec(t *testing.T) {
	tests := []struct {
		name    string
		section string
		want    ClusterSpec
		wantErr string
	}{
		{
			name:    "validators alone leaves every option at its zero value",
			section: "validators: 3\n",
			want:    ClusterSpec{Validators: 3},
		},
		{
			name:    "every option set",
			section: "validators: 1\noracle: true\ncode-submission-policy: inert\npkg-approver: user\nblock-max-gas: 20000000\n",
			want: ClusterSpec{
				Validators:           1,
				Oracle:               true,
				CodeSubmissionPolicy: "inert",
				PkgApprover:          "user",
				BlockMaxGas:          20_000_000,
			},
		},
		{
			name:    "blank lines and comments are ignored",
			section: "\n# how many\nvalidators: 4\n\n# and the oracle\noracle: true\n",
			want:    ClusterSpec{Validators: 4, Oracle: true},
		},
		{
			name:    "surrounding space is not part of the value",
			section: "  validators :  2  \n  oracle :  true  \n",
			want:    ClusterSpec{Validators: 2, Oracle: true},
		},
		{
			name:    "oracle defaults off, so it must be opted into",
			section: "validators: 1\noracle: false\n",
			want:    ClusterSpec{Validators: 1, Oracle: false},
		},
		{
			name:    "validators is the one key with no usable zero value",
			section: "oracle: true\n",
			wantErr: `cluster section: "validators" is required`,
		},
		{
			name:    "a cluster of no validators is not a cluster",
			section: "validators: 0\n",
			wantErr: `cluster section: validators must be at least 1, got 0`,
		},
		{
			name:    "a negative count is rejected rather than clamped",
			section: "validators: -1\n",
			wantErr: `cluster section: validators must be at least 1, got -1`,
		},
		{
			name:    "an unknown key is a typo, not something to ignore",
			section: "validators: 1\nvalidatorz: 2\n",
			wantErr: `cluster section: unknown key "validatorz"`,
		},
		{
			name:    "a non-numeric count names the offending value",
			section: "validators: three\n",
			wantErr: `cluster section: validators "three": invalid syntax`,
		},
		{
			name:    "a non-boolean flag names the offending value",
			section: "validators: 1\noracle: yes please\n",
			wantErr: `cluster section: oracle "yes please": invalid syntax`,
		},
		{
			name:    "a line with no separator cannot be a setting",
			section: "validators: 1\noracle\n",
			wantErr: `cluster section: expected "key: value", got "oracle"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseClusterSpec(script(tt.section))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// The section is required rather than optional, so that reading a scenario
// never leaves the cluster it runs against implicit.
func TestParseClusterSpecRequiresTheSection(t *testing.T) {
	_, err := ParseClusterSpec([]byte("# a scenario\ngnokey query auth/accounts\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `no "cluster" section`)
}

// Grouping scripts by cluster is what lets one boot serve several scenarios,
// so two scripts asking for the same thing must compare equal, and the spec
// has to be usable as a map key at all.
func TestClusterSpecIsComparable(t *testing.T) {
	a, err := ParseClusterSpec(script("validators: 3\noracle: true\n"))
	require.NoError(t, err)
	b, err := ParseClusterSpec(script("oracle: true\nvalidators: 3\n"))
	require.NoError(t, err)
	c, err := ParseClusterSpec(script("validators: 4\noracle: true\n"))
	require.NoError(t, err)

	counts := map[ClusterSpec]int{}
	counts[a]++
	counts[b]++
	counts[c]++

	assert.Equal(t, 2, counts[a], "key order must not change the spec")
	assert.Len(t, counts, 2, fmt.Sprintf("differing validator counts are different clusters: %v", counts))
}
