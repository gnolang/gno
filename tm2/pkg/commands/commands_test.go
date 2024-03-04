package commands

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/peterbourgon/ff/v3/fftest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type configDelegate func(*flag.FlagSet)

type mockConfig struct {
	configFn configDelegate
}

func (c *mockConfig) RegisterFlags(fs *flag.FlagSet) {
	if c.configFn != nil {
		c.configFn(fs)
	}
}

func TestCommandParseAndRun(t *testing.T) {
	t.Parallel()

	type flags struct {
		b bool
		s string
		x bool
	}
	tests := []struct {
		name          string
		osArgs        []string
		expectedCmd   string
		expectedArgs  []string
		expectedFlags flags
		expectedError string
	}{
		{
			name:          "no args no flags",
			expectedCmd:   "main",
			osArgs:        []string{},
			expectedArgs:  []string{},
			expectedFlags: flags{},
		},
		{
			name:          "only args",
			expectedCmd:   "main",
			osArgs:        []string{"bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{},
		},
		{
			name:          "only flags",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "-s", "str"},
			expectedArgs:  []string{},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "ignore all flags",
			expectedCmd:   "main",
			osArgs:        []string{"--", "-b", "-s", "str", "bar"},
			expectedArgs:  []string{"-b", "-s", "str", "bar"},
			expectedFlags: flags{},
		},
		{
			name:          "ignore some flags",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "--", "-s", "--", "str", "bar"},
			expectedArgs:  []string{"-s", "--", "str", "bar"},
			expectedFlags: flags{b: true},
		},
		{
			name:          "unknow flag",
			expectedCmd:   "main",
			osArgs:        []string{"-y", "-s", "str"},
			expectedArgs:  []string{},
			expectedError: "error parsing commandline arguments: flag provided but not defined: -y",
		},
		{
			name:          "flags before args",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "-s", "str", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags after args",
			expectedCmd:   "main",
			osArgs:        []string{"bar", "baz", "-b", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags around args",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "bar", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "flags between args",
			expectedCmd:   "main",
			osArgs:        []string{"bar", "-b", "-s", "str", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "ignore ending --",
			expectedCmd:   "main",
			osArgs:        []string{"bar", "-b", "-s", "str", "--"},
			expectedArgs:  []string{"bar"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "args and some ignored flags",
			expectedCmd:   "main",
			osArgs:        []string{"bar", "-b", "--", "-s", "--", "str", "baz"},
			expectedArgs:  []string{"bar", "-s", "--", "str", "baz"},
			expectedFlags: flags{b: true},
		},
		{
			name:          "subcommand no flags no args",
			expectedCmd:   "sub",
			osArgs:        []string{"sub"},
			expectedArgs:  []string{},
			expectedFlags: flags{},
		},
		{
			name:          "subcommand only args",
			expectedCmd:   "sub",
			osArgs:        []string{"sub", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{},
		},
		{
			name:          "subcommand flag before subcommand",
			expectedCmd:   "sub",
			osArgs:        []string{"-x", "sub"},
			expectedError: "error parsing commandline arguments: flag provided but not defined: -x",
		},
		{
			name:          "subcommand only flags",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "-x", "-s", "str"},
			expectedArgs:  []string{},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand ignore all flags after --",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "--", "-x", "-s", "str"},
			expectedArgs:  []string{"-x", "-s", "str"},
			expectedFlags: flags{b: true},
		},
		{
			name:          "subcommand ignore some flags after --",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "-x", "--", "-s", "str"},
			expectedArgs:  []string{"-s", "str"},
			expectedFlags: flags{b: true, x: true},
		},
		{
			name:          "subcommand ignored by --",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "--", "sub", "-x", "-s", "str"},
			expectedArgs:  []string{"sub", "-x", "-s", "str"},
			expectedFlags: flags{b: true},
		},
		{
			name:          "subcommand ignored by preceding arg",
			expectedCmd:   "main",
			osArgs:        []string{"-b", "bar", "sub", "-s", "str"},
			expectedArgs:  []string{"bar", "sub"},
			expectedFlags: flags{b: true, s: "str"},
		},
		{
			name:          "subcommand flags before args",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "-x", "-s", "str", "bar", "baz"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags after args",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "bar", "baz", "-x", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags around args",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "-x", "bar", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subcommand flags between args",
			expectedCmd:   "sub",
			osArgs:        []string{"-b", "sub", "bar", "-x", "baz", "-s", "str"},
			expectedArgs:  []string{"bar", "baz"},
			expectedFlags: flags{b: true, s: "str", x: true},
		},
		{
			name:          "subsubcommand with parent flags",
			expectedCmd:   "subsub",
			osArgs:        []string{"-b", "sub", "-x", "subsub", "bar"},
			expectedArgs:  []string{"bar"},
			expectedFlags: flags{b: true, x: true},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				invokedCmd string
				args       []string
				flags      flags
			)
			// Create a cmd main that takes 2 flags -b and -s
			cmd := NewCommand(
				Metadata{Name: "main"},
				&mockConfig{
					configFn: func(fs *flag.FlagSet) {
						fs.BoolVar(&flags.b, "b", false, "a boolan")
						fs.StringVar(&flags.s, "s", "", "a string")
					},
				},
				func(_ context.Context, a []string) error {
					invokedCmd = "main"
					args = a
					return nil
				},
			)
			// Add a sub command to cmd with a single flag -x
			subcmd := NewCommand(
				Metadata{Name: "sub"},
				&mockConfig{
					configFn: func(fs *flag.FlagSet) {
						fs.BoolVar(&flags.x, "x", false, "a boolan")
					},
				},
				func(_ context.Context, a []string) error {
					invokedCmd = "sub"
					args = a
					return nil
				},
			)
			cmd.AddSubCommands(subcmd)
			// Add a sub command to sub cmd
			subcmd.AddSubCommands(
				NewCommand(
					Metadata{Name: "subsub"},
					&mockConfig{
						configFn: func(fs *flag.FlagSet) {},
					},
					func(_ context.Context, a []string) error {
						invokedCmd = "subsub"
						args = a
						return nil
					},
				),
			)

			err := cmd.ParseAndRun(context.Background(), tt.osArgs)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedCmd, invokedCmd, "wrong cmd")
			require.Equal(t, tt.expectedArgs, args, "wrong args")
			require.Equal(t, tt.expectedFlags, flags, "wrong flags")
		})
	}
}

func TestCommand_AddSubCommands(t *testing.T) {
	t.Parallel()

	// Test setup //

	type testCmd struct {
		cmd     *Command
		subCmds []*testCmd
	}

	getSubcommands := func(t *testCmd) []*Command {
		res := make([]*Command, len(t.subCmds))

		for i, subCmd := range t.subCmds {
			res[i] = subCmd.cmd
		}

		return res
	}

	generateTestCmd := func(name string) *Command {
		return NewCommand(
			Metadata{
				Name: name,
			},
			&mockConfig{
				func(fs *flag.FlagSet) {
					fs.String(
						name,
						"",
						"",
					)
				},
			},
			HelpExec,
		)
	}

	var postorderCommands func(root *testCmd) []*testCmd

	postorderCommands = func(root *testCmd) []*testCmd {
		if root == nil {
			return nil
		}

		res := make([]*testCmd, 0)

		for _, child := range root.subCmds {
			res = append(res, postorderCommands(child)...)
		}

		return append(res, root)
	}

	// Cases //

	testTable := []struct {
		name   string
		topCmd *testCmd
	}{
		{
			name: "no subcommands",
			topCmd: &testCmd{
				cmd:     generateTestCmd("level0"),
				subCmds: nil,
			},
		},
		{
			name: "single subcommand level",
			topCmd: &testCmd{
				cmd: generateTestCmd("level0"),
				subCmds: []*testCmd{
					{
						cmd:     generateTestCmd("level1"),
						subCmds: nil,
					},
				},
			},
		},
		{
			name: "multiple subcommand levels",
			topCmd: &testCmd{
				cmd: generateTestCmd("level0"),
				subCmds: []*testCmd{
					{
						cmd: generateTestCmd("level1"),
						subCmds: []*testCmd{
							{
								cmd:     generateTestCmd("level2"),
								subCmds: nil,
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var validateSubcommandTree func(flag string, root *Command)

			validateSubcommandTree = func(flag string, root *Command) {
				assert.NotNil(t, root.flagSet.Lookup(flag))

				for _, subcommand := range root.subcommands {
					validateSubcommandTree(flag, subcommand)
				}
			}

			// Register the subcommands in LIFO order (postorder), starting from the
			// leaf of the command tree (mimics how the commands package is used)
			commandOrder := postorderCommands(testCase.topCmd)

			for _, currCmd := range commandOrder {
				// For the current command, register its subcommand tree
				currCmd.cmd.AddSubCommands(getSubcommands(currCmd)...)

				// Validate that the entire subcommand tree has root command flags
				for _, subCmd := range currCmd.cmd.subcommands {
					// For each root command flag, validate
					currCmd.cmd.flagSet.VisitAll(func(f *flag.Flag) {
						validateSubcommandTree(f.Name, subCmd)
					})
				}
			}
		})
	}
}

// Forked from peterbourgon/ff/ffcli
func TestHelpUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		command        *Command
		expectedOutput string
	}{
		{
			name: "normal case",
			command: &Command{
				name:       "TestHelpUsage",
				shortUsage: "TestHelpUsage [flags] <args>",
				shortHelp:  "some short help",
				longHelp:   "Some long help.",
			},
			expectedOutput: strings.TrimSpace(`
USAGE
  TestHelpUsage [flags] <args>

Some long help.

FLAGS
  -b=false  bool
  -d 0s     time.Duration
  -f 0      float64
  -i 0      int
  -s ...    string
  -x ...    collection of strings (repeatable)
`) + "\n\n",
		},
		{
			name: "no long help",
			command: &Command{
				name:       "TestHelpUsage",
				shortUsage: "TestHelpUsage [flags] <args>",
				shortHelp:  "some short help",
			},
			expectedOutput: strings.TrimSpace(`
USAGE
  TestHelpUsage [flags] <args>

some short help.

FLAGS
  -b=false  bool
  -d 0s     time.Duration
  -f 0      float64
  -i 0      int
  -s ...    string
  -x ...    collection of strings (repeatable)
`) + "\n\n",
		},
		{
			name: "no short and no long help",
			command: &Command{
				name:       "TestHelpUsage",
				shortUsage: "TestHelpUsage [flags] <args>",
			},
			expectedOutput: strings.TrimSpace(`
USAGE
  TestHelpUsage [flags] <args>

FLAGS
  -b=false  bool
  -d 0s     time.Duration
  -f 0      float64
  -i 0      int
  -s ...    string
  -x ...    collection of strings (repeatable)
`) + "\n\n",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fs, _ := fftest.Pair()
			var buf bytes.Buffer
			fs.SetOutput(&buf)

			tt.command.flagSet = fs

			err := tt.command.ParseAndRun(context.Background(), []string{"-h"})

			assert.ErrorIs(t, err, flag.ErrHelp)
			assert.Equal(t, tt.expectedOutput, buf.String())
		})
	}
}

// Forked from peterbourgon/ff/ffcli
func TestNestedOutput(t *testing.T) {
	t.Parallel()

	var (
		rootHelpOutput = "USAGE\n  \n\nSUBCOMMANDS\n  foo\n\n"
		fooHelpOutput  = "USAGE\n  foo\n\nSUBCOMMANDS\n  bar\n\n"
		barHelpOutout  = "USAGE\n  bar\n\n"
	)
	for _, testcase := range []struct {
		name       string
		args       []string
		wantErr    error
		wantOutput string
	}{
		{
			name:       "root without args",
			args:       []string{},
			wantErr:    flag.ErrHelp,
			wantOutput: rootHelpOutput,
		},
		{
			name:       "root with args",
			args:       []string{"abc", "def ghi"},
			wantErr:    flag.ErrHelp,
			wantOutput: rootHelpOutput,
		},
		{
			name:       "root help",
			args:       []string{"-h"},
			wantErr:    flag.ErrHelp,
			wantOutput: rootHelpOutput,
		},
		{
			name:       "foo without args",
			args:       []string{"foo"},
			wantOutput: "foo: ''\n",
		},
		{
			name:       "foo with args",
			args:       []string{"foo", "alpha", "beta"},
			wantOutput: "foo: 'alpha beta'\n",
		},
		{
			name:       "foo help",
			args:       []string{"foo", "-h"},
			wantErr:    flag.ErrHelp,
			wantOutput: fooHelpOutput, // only one instance of usage string
		},
		{
			name:       "foo bar without args",
			args:       []string{"foo", "bar"},
			wantErr:    flag.ErrHelp,
			wantOutput: barHelpOutout,
		},
		{
			name:       "foo bar with args",
			args:       []string{"foo", "bar", "--", "baz quux"},
			wantErr:    flag.ErrHelp,
			wantOutput: barHelpOutout,
		},
		{
			name:       "foo bar help",
			args:       []string{"foo", "bar", "--help"},
			wantErr:    flag.ErrHelp,
			wantOutput: barHelpOutout,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			t.Parallel()

			var (
				rootfs = flag.NewFlagSet("root", flag.ContinueOnError)
				foofs  = flag.NewFlagSet("foo", flag.ContinueOnError)
				barfs  = flag.NewFlagSet("bar", flag.ContinueOnError)
				buf    bytes.Buffer
			)
			rootfs.SetOutput(&buf)
			foofs.SetOutput(&buf)
			barfs.SetOutput(&buf)

			barExec := func(_ context.Context, args []string) error {
				return flag.ErrHelp
			}

			bar := &Command{
				name:    "bar",
				flagSet: barfs,
				exec:    barExec,
			}

			fooExec := func(_ context.Context, args []string) error {
				fmt.Fprintf(&buf, "foo: '%s'\n", strings.Join(args, " "))
				return nil
			}

			foo := &Command{
				name:        "foo",
				flagSet:     foofs,
				subcommands: []*Command{bar},
				exec:        fooExec,
			}

			rootExec := func(_ context.Context, args []string) error {
				return flag.ErrHelp
			}

			root := &Command{
				flagSet:     rootfs,
				subcommands: []*Command{foo},
				exec:        rootExec,
			}

			err := root.ParseAndRun(context.Background(), testcase.args)

			assert.ErrorIs(t, err, testcase.wantErr)
			assert.Equal(t, testcase.wantOutput, buf.String())
		})
	}
}
