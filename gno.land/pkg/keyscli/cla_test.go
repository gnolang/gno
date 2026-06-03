package keyscli

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	tmerrors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestIsCLAError(t *testing.T) {
	// Real CLA error: UnauthorizedUserError wrapped with the CLA log.
	claErr := tmerrors.Wrapf(vm.UnauthorizedUserError{}, "deliver transaction failed: log:address g1abc has not signed the required CLA")
	assert.True(t, isCLAError(claErr))
	// Extra wrap (as maketx.go does with "broadcast tx") must still match.
	assert.True(t, isCLAError(tmerrors.Wrap(claErr, "broadcast tx")))

	// Namespace error: same UnauthorizedUserError cause, no CLA substring.
	nsErr := tmerrors.Wrapf(vm.UnauthorizedUserError{}, "deliver transaction failed: log:user g1abc is not allowed to deploy to namespace")
	assert.False(t, isCLAError(nsErr))

	// Plain errors whose Cause is not UnauthorizedUserError must not match,
	// even if the string happens to contain the CLA substring.
	assert.False(t, isCLAError(fmt.Errorf("address g1abc has not signed the required CLA")))
	assert.False(t, isCLAError(fmt.Errorf("unauthorized user")))
	assert.False(t, isCLAError(fmt.Errorf("")))
	assert.False(t, isCLAError(nil))
}

func TestParseQEvalString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"quoted hash", `("abc123hash" string)`, "abc123hash", true},
		{"quoted url", `("https://example.com/cla" string)`, "https://example.com/cla", true},
		{"empty quoted string is valid", `("" string)`, "", true},
		{"quoted with spaces", `("hello world" string)`, "hello world", true},
		{"quoted with escapes", `("\"quoted\" value" string)`, `"quoted" value`, true},
		{"unquoted fallback", "(abc123hash string)", "abc123hash", true},

		{"non-string type", "(true bool)", "", false},
		{"empty input", "", "", false},
		{"garbage", "garbage", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseQEvalString(tc.in)
			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormatCLAHelper(t *testing.T) {
	helper := formatCLAHelper("abc123hash", "https://example.com/cla", "gno.land/r/sys/cla", "testchain", "localhost:26657", "g1abc")

	assert.Contains(t, helper, "CLA document: https://example.com/cla")
	assert.Contains(t, helper, "To sign the CLA, run:")
	assert.Contains(t, helper, "-pkgpath gno.land/r/sys/cla")
	assert.Contains(t, helper, "-func Sign")
	assert.Contains(t, helper, "-args abc123hash")
	assert.Contains(t, helper, "-chainid testchain")
	assert.Contains(t, helper, "g1abc")
}

func TestFormatCLAHelper_NoHash(t *testing.T) {
	helper := formatCLAHelper("", "https://example.com", "gno.land/r/sys/cla", "chain", "localhost:26657", "g1abc")
	assert.Contains(t, helper, "<CLA_HASH>")
	assert.Contains(t, helper, "To sign the CLA, run:")
}

func TestFormatCLAHelper_NoURL(t *testing.T) {
	helper := formatCLAHelper("abc123", "", "gno.land/r/sys/cla", "chain", "localhost:26657", "g1abc")
	assert.NotContains(t, helper, "CLA document:")
	assert.Contains(t, helper, "To sign the CLA, run:")
}

func TestFormatCLAHelper_NoChainIDNoRemote(t *testing.T) {
	helper := formatCLAHelper("abc123", "", "gno.land/r/sys/cla", "", "", "g1abc")
	assert.NotContains(t, helper, "-chainid")
	assert.NotContains(t, helper, "-remote")
}
