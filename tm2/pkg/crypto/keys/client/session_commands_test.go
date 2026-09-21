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
func runSessionCmd(kbHome string, args ...string) (string, error) {
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

// sessionTest is a single table entry. baseArgs builds the minimal valid
// invocation against the env's master/session keys; mutate (optional) modifies
// the args; wantStdout (optional) lists substrings expected in stdout. The
// function-of-env signatures let each subtest run in parallel against its own
// fresh keybase.
type sessionTest struct {
	name        string
	baseArgs    func(env sessionTestEnv) []string
	mutate      func([]string) []string
	wantErr     error
	wantErrText string
	wantStdout  func(env sessionTestEnv) []string
}

// runSessionTests drives each test in parallel against its own fresh keybase
// (to avoid leveldb lock contention while keeping outer t.Parallel valid).
func runSessionTests(t *testing.T, tests []sessionTest) {
	t.Helper()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			env := setupSessionTestEnv(t)
			args := tc.baseArgs(env)
			if tc.mutate != nil {
				args = tc.mutate(args)
			}
			out, err := runSessionCmd(env.kbHome, args...)
			switch {
			case tc.wantErr != nil:
				assert.ErrorIs(t, err, tc.wantErr)
			case tc.wantErrText != "":
				assert.ErrorContains(t, err, tc.wantErrText)
			default:
				require.NoError(t, err)
			}
			if tc.wantStdout != nil {
				for _, want := range tc.wantStdout(env) {
					assert.Contains(t, out, want)
				}
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
	_, err := runSessionCmd(env.kbHome, "maketx", "session")
	assert.ErrorIs(t, err, flag.ErrHelp)
}

// createBaseArgs / revokeBaseArgs / revokeAllBaseArgs build the minimal valid
// invocation for each session subcommand. Used by sessionTest entries; each
// subtest gets a fresh env so these are evaluated per-subtest (parallel-safe).
func createBaseArgs(env sessionTestEnv) []string {
	return []string{
		"maketx", "session", "create",
		"--gas-wanted", "100000",
		"--gas-fee", "1ugnot",
		"--pubkey", env.sessionPubkey,
		"--expires-at", "24h",
		"--allow-paths", "*",
		"--broadcast=false",
		env.masterName,
	}
}

func revokeBaseArgs(env sessionTestEnv) []string {
	return []string{
		"maketx", "session", "revoke",
		"--gas-wanted", "100000",
		"--gas-fee", "1ugnot",
		"--pubkey", env.sessionPubkey,
		"--broadcast=false",
		env.masterName,
	}
}

func revokeAllBaseArgs(env sessionTestEnv) []string {
	return []string{
		"maketx", "session", "revokeall",
		"--gas-wanted", "100000",
		"--gas-fee", "1ugnot",
		"--broadcast=false",
		env.masterName,
	}
}

// addMasterFlag mutates args to append "--master <some-name>". The actual name
// doesn't matter — the CLI rejects --master on session-lifecycle commands
// before resolving the name.
func addMasterFlag(a []string) []string {
	return append(a, "--master", "anymaster")
}

func TestSession_Create(t *testing.T) {
	t.Parallel()
	runSessionTests(t, []sessionTest{
		{
			name:     "happy path with wildcard allow-paths",
			baseArgs: createBaseArgs,
			wantStdout: func(env sessionTestEnv) []string {
				return []string{`"@type":"/auth.m_create_session"`, env.masterAddr, `"expires_at":`, `"allow_paths":["*"]`}
			},
		},
		{
			name:        "missing --allow-paths rejected",
			baseArgs:    createBaseArgs,
			mutate:      dropFlag("--allow-paths"),
			wantErrText: "--allow-paths is required",
		},
		{
			name:        "wildcard with path rejected",
			baseArgs:    createBaseArgs,
			mutate:      setFlag("--allow-paths", "*:gno.land/r/foo"),
			wantErrText: "wildcard '*' must not have a path suffix",
		},
		{
			name:     "no positional arg returns help",
			baseArgs: createBaseArgs,
			mutate:   dropPositional,
			wantErr:  flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			baseArgs:    createBaseArgs,
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing gas-fee",
			baseArgs:    createBaseArgs,
			mutate:      dropFlag("--gas-fee"),
			wantErrText: "gas-fee not specified",
		},
		{
			name:        "missing pubkey",
			baseArgs:    createBaseArgs,
			mutate:      dropFlag("--pubkey"),
			wantErrText: "pubkey must be specified",
		},
		{
			name:     "negative spend-period",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--spend-period", "-1")
			},
			wantErrText: "spend-period must be non-negative",
		},
		{
			name:        "expires-at required",
			baseArgs:    createBaseArgs,
			mutate:      dropFlag("--expires-at"),
			wantErrText: "--expires-at is required",
		},
		{
			name:        "expires-at exceeds cap",
			baseArgs:    createBaseArgs,
			mutate:      setFlag("--expires-at", "1461d"),
			wantErrText: "exceeds chain max",
		},
		{
			name:       "expires-at none accepted",
			baseArgs:   createBaseArgs,
			mutate:     setFlag("--expires-at", "none"),
			wantStdout: func(_ sessionTestEnv) []string { return []string{`"@type":"/auth.m_create_session"`} },
		},
		{
			name:        "invalid pubkey bech32",
			baseArgs:    createBaseArgs,
			mutate:      setFlag("--pubkey", "not-a-pubkey"),
			wantErrText: "unable to parse public key from bech32",
		},
		{
			name:     "non-existent master key",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				a[len(a)-1] = "nonexistent"
				return a
			},
			wantErrText: "Key nonexistent not found",
		},
		{
			name:     "spend-limit and allow-paths flow into JSON",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				a = setFlag("--allow-paths", "vm/exec:gno.land/r/foo")(a)
				return append(a, "--spend-limit", "1000ugnot")
			},
			wantStdout: func(_ sessionTestEnv) []string {
				return []string{`"spend_limit":"1000ugnot"`, `"allow_paths":["vm/exec:gno.land/r/foo"]`}
			},
		},
		{
			name:     "multiple allow-paths accumulate",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				a = setFlag("--allow-paths", "vm/exec:gno.land/r/a")(a)
				return append(a, "--allow-paths", "bank/send")
			},
			wantStdout: func(_ sessionTestEnv) []string {
				return []string{`"allow_paths":["vm/exec:gno.land/r/a","bank/send"]`}
			},
		},
		{
			name:     "empty allow-paths entry rejected",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--allow-paths", "")
			},
			wantErrText: "entry is empty",
		},
		{
			name:     "allow-paths trailing slash rejected",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--allow-paths", "vm/exec:gno.land/r/foo/")
			},
			wantErrText: "must not end with /",
		},
		{
			name:     "bare 'bank' rejected (missing /)",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--allow-paths", "bank")
			},
			wantErrText: "<route>/<type>",
		},
		{
			name:     "empty path after colon rejected",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--allow-paths", "vm/exec:")
			},
			wantErrText: "path after ':' must be non-empty",
		},
		{
			name:     "extra slash in route_type rejected",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--allow-paths", "vm/exec/extra")
			},
			wantErrText: "<route>/<type>",
		},
		{
			name:     "explicit spend-period 0 with spend-limit (lifetime cap)",
			baseArgs: createBaseArgs,
			mutate: func(a []string) []string {
				return append(a, "--spend-limit", "1000ugnot", "--spend-period", "0")
			},
			// SpendPeriod has json:",omitempty" so 0 is omitted; SpendLimit appears.
			wantStdout: func(_ sessionTestEnv) []string { return []string{`"spend_limit":"1000ugnot"`} },
		},
		{
			name:        "expires-at 0 rejected",
			baseArgs:    createBaseArgs,
			mutate:      setFlag("--expires-at", "0"),
			wantErrText: "must be a positive duration",
		},
		{
			name:        "expires-at bare integer rejected (missing unit)",
			baseArgs:    createBaseArgs,
			mutate:      setFlag("--expires-at", "24"),
			wantErrText: "must be a future unix timestamp",
		},
		{
			name:        "--master rejected on session create",
			baseArgs:    createBaseArgs,
			mutate:      addMasterFlag,
			wantErrText: "--master cannot be used with session create/revoke/revokeall",
		},
	})
}

func TestSession_Revoke(t *testing.T) {
	t.Parallel()
	runSessionTests(t, []sessionTest{
		{
			name:     "happy path prints unsigned tx JSON",
			baseArgs: revokeBaseArgs,
			wantStdout: func(env sessionTestEnv) []string {
				return []string{`"@type":"/auth.m_revoke_session"`, env.masterAddr}
			},
		},
		{
			name:     "no positional arg returns help",
			baseArgs: revokeBaseArgs,
			mutate:   dropPositional,
			wantErr:  flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			baseArgs:    revokeBaseArgs,
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing pubkey",
			baseArgs:    revokeBaseArgs,
			mutate:      dropFlag("--pubkey"),
			wantErrText: "pubkey must be specified",
		},
		{
			name:        "invalid pubkey bech32",
			baseArgs:    revokeBaseArgs,
			mutate:      setFlag("--pubkey", "not-a-pubkey"),
			wantErrText: "unable to parse public key from bech32",
		},
		{
			name:        "--master rejected on session revoke",
			baseArgs:    revokeBaseArgs,
			mutate:      addMasterFlag,
			wantErrText: "--master cannot be used with session create/revoke/revokeall",
		},
	})
}

func TestSession_RevokeAll(t *testing.T) {
	t.Parallel()
	runSessionTests(t, []sessionTest{
		{
			name:     "happy path prints unsigned tx JSON",
			baseArgs: revokeAllBaseArgs,
			wantStdout: func(env sessionTestEnv) []string {
				return []string{`"@type":"/auth.m_revoke_all_sessions"`, env.masterAddr}
			},
		},
		{
			name:     "no positional arg returns help",
			baseArgs: revokeAllBaseArgs,
			mutate:   dropPositional,
			wantErr:  flag.ErrHelp,
		},
		{
			name:        "missing gas-wanted",
			baseArgs:    revokeAllBaseArgs,
			mutate:      dropFlag("--gas-wanted"),
			wantErrText: "gas-wanted not specified",
		},
		{
			name:        "missing gas-fee",
			baseArgs:    revokeAllBaseArgs,
			mutate:      dropFlag("--gas-fee"),
			wantErrText: "gas-fee not specified",
		},
		{
			name:        "--master rejected on session revokeall",
			baseArgs:    revokeAllBaseArgs,
			mutate:      addMasterFlag,
			wantErrText: "--master cannot be used with session create/revoke/revokeall",
		},
	})
}
