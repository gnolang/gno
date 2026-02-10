package main

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanApp(t *testing.T) {
	tc := []testMainCase{
		{
			args:        []string{"clean -h"},
			errShouldBe: "flag: help requested",
		},
		{
			args:        []string{"clean unknown"},
			errShouldBe: "flag: help requested",
		},
		{
			args:                 []string{"clean"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"clean"},
			testDir:              "../../tests/integ/minimalist_gnomod",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"clean", "-modcache"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
		},
		{
			args:                 []string{"clean", "-modcache", "-n"},
			testDir:              "../../tests/integ/empty_dir",
			simulateExternalRepo: true,
			stdoutShouldContain:  "rm -rf ",
		},
	}
	testMainCaseRun(t, tc)

	workingDir, err := os.Getwd()
	require.NoError(t, err)

	// Test clean command
	for _, tc := range []struct {
		desc         string
		args         []string
		files        []string
		filesRemoved []string
		stdOut       string
	}{
		{
			desc:         "only_generated_files",
			args:         []string{"clean"},
			files:        []string{"gno.mod", "tmp.gno.gen.go", "tmp.gno.gen_test.go"},
			filesRemoved: []string{"tmp.gno.gen.go", "tmp.gno.gen_test.go"},
			stdOut:       "rm tmp.gno.gen.go\nrm tmp.gno.gen_test.go\n",
		},
		{
			desc:  "no_generated_files",
			args:  []string{"clean"},
			files: []string{"gno.mod", "tmp.gno"},
		},
		{
			desc:         "mixed_files",
			args:         []string{"clean"},
			files:        []string{"gno.mod", "README.md", "tmp.gno", "tmp_test.gno", "tmp.gno.gen.go", "tmp.gno.gen_test.go"},
			filesRemoved: []string{"tmp.gno.gen.go", "tmp.gno.gen_test.go"},
			stdOut:       "rm tmp.gno.gen.go\nrm tmp.gno.gen_test.go\n",
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")

			// test clean with external repo
			tmpDir, cleanUpFn := createTmpDir(t)
			defer cleanUpFn()

			// cd to tmp directory
			os.Chdir(tmpDir)
			defer os.Chdir(workingDir)

			// create files
			for _, file := range tc.files {
				err = os.WriteFile(file, []byte("test"), 0o644)
				require.NoError(t, err)
			}

			// set up io
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			// dry run clean
			cmd, _ := newGnocliCmd(io)
			err = cmd.ParseAndRun(context.Background(), []string{"clean", "-n"})
			require.NoError(t, err)
			// check output
			if tc.stdOut != "" {
				assert.Equal(t, tc.stdOut, mockOut.String())
			}
			// check files
			for _, file := range tc.files {
				_, err = os.Stat(file)
				require.NoError(t, err)
			}

			// run clean
			cmd, _ = newGnocliCmd(io)
			err = cmd.ParseAndRun(context.Background(), []string{"clean"})
			require.NoError(t, err)
			// check files
			for _, file := range tc.filesRemoved {
				_, err = os.Stat(file)
				assert.True(t, os.IsNotExist(err), "expected: ErrNotExist, got: %v", err)
			}
		})
	}
}
