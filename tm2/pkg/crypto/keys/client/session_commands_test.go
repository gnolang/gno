package client

import (
	"bytes"
	"context"
	"flag"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/crypto/keys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sessionTestEnv holds shared state for session-command tests: a temp keybase
// with a single master key, plus a freshly-generated session pubkey to pass
// as --pubkey on create/revoke.
type sessionTestEnv struct {
	kbHome        string
	masterName    string
	masterAddr    string
	sessionPubkey string
}

func setupSessionTestEnv(t *testing.T) sessionTestEnv {
	t.Helper()
	kbHome := t.TempDir()
	kb, err := keys.NewKeyBaseFromDir(kbHome)
	require.NoError(t, err)

	masterName := "master"
	masterInfo, err := kb.CreateAccount(masterName, generateTestMnemonic(t), "", "pw", 0, 0)
	require.NoError(t, err)

	sessionInfo, err := kb.CreateAccount("session-tmp", generateTestMnemonic(t), "", "pw", 0, 0)
	require.NoError(t, err)

	return sessionTestEnv{
		kbHome:        kbHome,
		masterName:    masterName,
		masterAddr:    crypto.AddressToBech32(masterInfo.GetAddress()),
		sessionPubkey: crypto.PubKeyToBech32(sessionInfo.GetPubKey()),
	}
}

// runSessionCmd runs `gnokey ...` against the test keybase and captures stdout.
// With --broadcast=false the command path never reaches sign/broadcast, so no
// password prompt is needed.
//
// Note: ParseAndRun must be called before reading out.String(), because Go
// evaluates a multi-return statement left-to-right and the buffer is empty
// until ParseAndRun finishes.
func runSessionCmd(t *testing.T, kbHome string, args ...string) (string, error) {
	t.Helper()
	io := commands.NewTestIO()
	out := &bytes.Buffer{}
	io.SetOut(commands.WriteNopCloser(out))
	io.SetErr(commands.WriteNopCloser(&bytes.Buffer{}))

	cmd := NewRootCmdWithBaseConfig(io, BaseOptions{
		InsecurePasswordStdin: true,
		Home:                  kbHome,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cmd.ParseAndRun(ctx, args)
	return out.String(), err
}

// sessionTest is a single table entry. mutate may be nil for the unmodified
// happy-path case.
type sessionTest struct {
	name        string
	mutate      func([]string) []string
	wantErr     error
	wantErrText string
	wantStdout  []string
}

// runSessionTests drives the tests against env.kbHome. Subtests run sequentially
// to avoid leveldb lock contention on the shared keybase.
func runSessionTests(t *testing.T, env sessionTestEnv, baseArgs func() []string, tests []sessionTest) {
	t.Helper()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := baseArgs()
			if tc.mutate != nil {
				args = tc.mutate(args)
			}
			out, err := runSessionCmd(t, env.kbHome, args...)
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrText != "":
				assert.ErrorContains(t, err, tc.wantErrText)
			default:
				require.NoError(t, err)
			}
			for _, want := range tc.wantStdout {
				assert.Contains(t, out, want)
			}
		})
	}
}

// dropFlag returns a mutator that removes `--name <value>` from args.
func dropFlag(name string) func([]string) []string {
	return func(args []string) []string {
		out := make([]string, 0, len(args))
		for i := 0; i < len(args); i++ {
			if args[i] == name {
				i++ // skip value
				continue
			}
			out = append(out, args[i])
		}
		return out
	}
}

// setFlag returns a mutator that replaces (or appends) `--name <value>` in args.
func setFlag(name, value string) func([]string) []string {
	return func(args []string) []string {
		out := make([]string, 0, len(args))
		replaced := false
		for i := 0; i < len(args); i++ {
			if args[i] == name && i+1 < len(args) {
				out = append(out, args[i], value)
				i++
				replaced = true
				continue
			}
			out = append(out, args[i])
		}
		if !replaced {
			out = append(out, name, value)
		}
		return out
	}
}

// dropPositional drops the last (positional) arg.
func dropPositional(args []string) []string {
	return args[:len(args)-1]
}

func TestSession_HelpExec(t *testing.T) {
	t.Parallel()
	env := setupSessionTestEnv(t)

	// `maketx session` with no subcommand should print help and return
	// flag.ErrHelp (matches every other parent-only dispatching command).
	_, err := runSessionCmd(t, env.kbHome, "maketx", "session")
	assert.ErrorIs(t, err, flag.ErrHelp)
}

func TestSession_Create(t *testing.T) {
	t.Parallel()
	env := setupSessionTestEnv(t)

	baseArgs := func() []string {
		return []string{
			"maketx", "session", "create",
			"--gas-wanted", "100000",
			"--gas-fee", "1ugnot",
			"--pubkey", env.sessionPubkey,
			"--expires-at", "24h",
			"--broadcast=false",
			env.masterName,
		}
	}

	runSessionTests(t, env, baseArgs, []sessionTest{
		{
			name:       "happy path prints unsigned tx JSON",
			wantStdout: []string{`"@type":"/auth.m_create_session"`, env.masterAddr, `"expires_at":`},
		},
		{
			name:    "no positional arg returns help",
			mutate:  dropPositional,
			wantErr: flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing gas-fee",
			mutate:      dropFlag("--gas-fee"),
			wantErrText: "gas-fee not specified",
		},
		{
			name:        "missing pubkey",
			mutate:      dropFlag("--pubkey"),
			wantErrText: "pubkey must be specified",
		},
		{
			name: "negative spend-period",
			mutate: func(a []string) []string {
				return append(a, "--spend-period", "-1")
			},
			wantErrText: "spend-period must be non-negative",
		},
		{
			name:        "expires-at required",
			mutate:      dropFlag("--expires-at"),
			wantErrText: "--expires-at is required",
		},
		{
			name:        "expires-at exceeds cap",
			mutate:      setFlag("--expires-at", "1461d"),
			wantErrText: "exceeds chain max",
		},
		{
			name:       "expires-at none accepted",
			mutate:     setFlag("--expires-at", "none"),
			wantStdout: []string{`"@type":"/auth.m_create_session"`},
		},
		{
			name:        "invalid pubkey bech32",
			mutate:      setFlag("--pubkey", "not-a-pubkey"),
			wantErrText: "unable to parse public key from bech32",
		},
		{
			name: "non-existent master key",
			mutate: func(a []string) []string {
				a[len(a)-1] = "nonexistent"
				return a
			},
			wantErrText: "Key nonexistent not found",
		},
		{
			name: "spend-limit and allow-paths flow into JSON",
			mutate: func(a []string) []string {
				return append(a, "--spend-limit", "1000ugnot", "--allow-paths", "gno.land/r/foo")
			},
			wantStdout: []string{
				`"spend_limit":"1000ugnot"`,
				`"allow_paths":["gno.land/r/foo"]`,
			},
		},
	})
}

func TestSession_Revoke(t *testing.T) {
	t.Parallel()
	env := setupSessionTestEnv(t)

	baseArgs := func() []string {
		return []string{
			"maketx", "session", "revoke",
			"--gas-wanted", "100000",
			"--gas-fee", "1ugnot",
			"--pubkey", env.sessionPubkey,
			"--broadcast=false",
			env.masterName,
		}
	}

	runSessionTests(t, env, baseArgs, []sessionTest{
		{
			name:       "happy path prints unsigned tx JSON",
			wantStdout: []string{`"@type":"/auth.m_revoke_session"`, env.masterAddr},
		},
		{
			name:    "no positional arg returns help",
			mutate:  dropPositional,
			wantErr: flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing pubkey",
			mutate:      dropFlag("--pubkey"),
			wantErrText: "pubkey must be specified",
		},
		{
			name:        "invalid pubkey bech32",
			mutate:      setFlag("--pubkey", "not-a-pubkey"),
			wantErrText: "unable to parse public key from bech32",
		},
	})
}

func TestSession_RevokeAll(t *testing.T) {
	t.Parallel()
	env := setupSessionTestEnv(t)

	baseArgs := func() []string {
		return []string{
			"maketx", "session", "revokeall",
			"--gas-wanted", "100000",
			"--gas-fee", "1ugnot",
			"--broadcast=false",
			env.masterName,
		}
	}

	runSessionTests(t, env, baseArgs, []sessionTest{
		{
			name:       "happy path prints unsigned tx JSON",
			wantStdout: []string{`"@type":"/auth.m_revoke_all_sessions"`, env.masterAddr},
		},
		{
			name:    "no positional arg returns help",
			mutate:  dropPositional,
			wantErr: flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing gas-fee",
			mutate:      dropFlag("--gas-fee"),
			wantErrText: "gas-fee not specified",
		},
	})
}
