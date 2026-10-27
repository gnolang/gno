package client

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testProof is the proof of "hello" on gnoland-1, for testOrigin and testPurpose,
// by the test1 key, signed by @cosmjs/amino (see tm2/pkg/std/signdata_test.go).
const testProof = `{
  "signed": {
    "chain_id": "gnoland-1", "account_number": "0", "sequence": "0",
    "fee": {"amount": [], "gas": "0"},
    "msgs": [{"type": "sign/MsgSignData", "value": {"signer": "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5",
      "data": "hello", "origin": "https://gno.land", "purpose": "auth/SignIn"}}],
    "memo": ""
  },
  "signature": {
    "pubKey": {"typeUrl": "/tm.PubKeySecp256k1", "value": "A+FhNtsXHjLfSJk1lB8FbiL4mGPjc50Kt81J7EKDnJ2y"},
    "signature": "/MurcSWJpSzxdBDx+lwL4Ldo/n2J7NQ7vywYD1/qj7MSZ0Ms5QJ1E/xami0mU/EETYdC5BQDGJCUf/EhHJSAdw=="
  }
}`

const testProofSigner = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

func writeProof(t *testing.T, proof string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "proof.json")
	require.NoError(t, os.WriteFile(path, []byte(proof), 0o644))

	return path
}

func TestVerifyData(t *testing.T) {
	t.Parallel()

	t.Run("no address provided", func(t *testing.T) {
		t.Parallel()

		_, err := runKeysCmd(t, t.TempDir(), "", "verify-data",
			"--chainid", "gnoland-1", "--data", "hello", "--proof-path", writeProof(t, testProof))
		assert.ErrorIs(t, err, flag.ErrHelp)
	})

	t.Run("accepts a proof from another implementation", func(t *testing.T) {
		t.Parallel()

		out, err := runKeysCmd(t, t.TempDir(), "", "verify-data",
			"--chainid", "gnoland-1", "--data", "hello", "--proof-path", writeProof(t, testProof), testProofSigner)
		require.NoError(t, err)
		assert.Contains(t, string(out), testProofSigner)
	})

	t.Run("reads the proof from stdin", func(t *testing.T) {
		t.Parallel()

		_, err := runKeysCmd(t, t.TempDir(), testProof, "verify-data",
			"--chainid", "gnoland-1", "--data", "hello", "--proof-path", "-", testProofSigner)
		require.NoError(t, err)
	})

	// The file is compared as it is, a final newline included.
	t.Run("compares a data file as it is", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "data.txt")
		require.NoError(t, os.WriteFile(path, []byte("hello\n"), 0o644))

		_, err := runKeysCmd(t, t.TempDir(), "", "verify-data",
			"--chainid", "gnoland-1", "--data-path", path, "--proof-path", writeProof(t, testProof), testProofSigner)
		assert.ErrorContains(t, err, "signs other data")
	})

	t.Run("verifies a proof from sign-data, by key name", func(t *testing.T) {
		t.Parallel()

		st := newSignDataTest(t)
		path := filepath.Join(t.TempDir(), "proof.json")

		_, err := st.run(t, "--chainid", "dev", "--data", "challenge", "--output-document", path, st.keyName)
		require.NoError(t, err)

		out, err := runKeysCmd(t, st.kbHome, "", "verify-data",
			"--chainid", "dev", "--data", "challenge", "--proof-path", path, st.keyName)
		require.NoError(t, err)
		assert.Contains(t, string(out), st.address.String())
	})

	// Without the expectations, a valid signature over another chain, an old
	// challenge or someone else's address would pass for a proof. Each case
	// runs on testProof unless it gives its own.
	for _, testCase := range []struct {
		name    string
		args    []string
		proof   string
		wantErr string
	}{
		{
			name:    "no chain id",
			args:    []string{"--data", "hello", testProofSigner},
			wantErr: "chainid",
		},
		{
			name:    "no data",
			args:    []string{"--chainid", "gnoland-1", testProofSigner},
			wantErr: "no data",
		},
		{
			name:    "no origin",
			args:    []string{"--chainid", "gnoland-1", "--origin", "", "--data", "hello", testProofSigner},
			wantErr: "-origin",
		},
		{
			name:    "no purpose",
			args:    []string{"--chainid", "gnoland-1", "--purpose", "", "--data", "hello", testProofSigner},
			wantErr: "-purpose",
		},
		{
			name:    "another origin",
			args:    []string{"--chainid", "gnoland-1", "--origin", "https://evil.example", "--data", "hello", testProofSigner},
			wantErr: `for origin "https://gno.land", not "https://evil.example"`,
		},
		{
			name:    "an expected origin that is not one",
			args:    []string{"--chainid", "gnoland-1", "--origin", "https://gno.land/", "--data", "hello", testProofSigner},
			wantErr: "is not scheme://host[:port]",
		},
		{
			name:    "another purpose",
			args:    []string{"--chainid", "gnoland-1", "--purpose", "auth/Other", "--data", "hello", testProofSigner},
			wantErr: `for purpose "auth/SignIn", not "auth/Other"`,
		},
		{
			name:    "unknown key name",
			args:    []string{"--chainid", "gnoland-1", "--data", "hello", "TotallyExistingKey"},
			wantErr: "TotallyExistingKey",
		},
		{
			name:    "another chain",
			args:    []string{"--chainid", "test12", "--data", "hello", testProofSigner},
			wantErr: `for chain "gnoland-1", not "test12"`,
		},
		{
			name:    "other data",
			args:    []string{"--chainid", "gnoland-1", "--data", "hello again", testProofSigner},
			wantErr: "signs other data",
		},
		{
			name:    "another address",
			args:    []string{"--chainid", "gnoland-1", "--data", "hello", "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj"},
			wantErr: "g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj",
		},
		{
			name:    "tampered signature",
			args:    []string{"--chainid", "gnoland-1", "--data", "hello", testProofSigner},
			proof:   strings.Replace(testProof, `"/MurcS`, `"AMurcS`, 1),
			wantErr: "invalid sign-data signature",
		},
		{
			name:    "not a proof",
			args:    []string{"--chainid", "gnoland-1", "--data", "hello", testProofSigner},
			proof:   "hello",
			wantErr: "unable to read sign-data proof",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			proof := testCase.proof
			if proof == "" {
				proof = testProof
			}

			args := append([]string{"verify-data", "--proof-path", writeProof(t, proof)}, testCase.args...)
			_, err := runKeysCmd(t, t.TempDir(), "", args...)
			assert.ErrorContains(t, err, testCase.wantErr)
		})
	}

	t.Run("no proof", func(t *testing.T) {
		t.Parallel()

		_, err := runKeysCmd(t, t.TempDir(), "", "verify-data",
			"--chainid", "gnoland-1", "--data", "hello", testProofSigner)
		assert.ErrorContains(t, err, "proof-path")
	})
}
