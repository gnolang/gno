package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	tc := []testMainCase{
		{args: []string{""}, errShouldBe: "unknown command "},
	}
	testMainCaseRun(t, tc)
}

type testMainCase struct {
	args                 []string
	testDir              string
	simulateExternalRepo bool

	// for the following FooContain+FooBe expected couples, if both are empty,
	// then the test suite will require that the "got" is not empty.
	errShouldContain     string
	errShouldBe          string
	stderrShouldContain  string
	stdoutShouldBe       string
	stdoutShouldContain  string
	stderrShouldBe       string
	recoverShouldContain string
	recoverShouldBe      string
}

func testMainCaseRun(t *testing.T, tc []testMainCase) {
	t.Helper()

	workingDir, err := os.Getwd()
	require.Nil(t, err)

	for _, test := range tc {
		errShouldBeEmpty := test.errShouldContain == "" && test.errShouldBe == ""
		stdoutShouldBeEmpty := test.stdoutShouldContain == "" && test.stdoutShouldBe == ""
		stderrShouldBeEmpty := test.stderrShouldContain == "" && test.stderrShouldBe == ""
		recoverShouldBeEmpty := test.recoverShouldContain == "" && test.recoverShouldBe == ""

		testName := strings.Join(test.args, " ")
		testName = strings.ReplaceAll(testName+test.testDir, "/", "~")

		t.Run(testName, func(t *testing.T) {
			cmd := command.NewMockCommand()
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")
			stdout := command.WriteNopCloser(mockOut)
			stderr := command.WriteNopCloser(mockErr)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			require.NotNil(t, cmd)

			checkOutputs := func(t *testing.T) {
				t.Helper()

				if stdoutShouldBeEmpty {
					require.Empty(t, mockOut.String(), "stdout should be empty")
				} else {
					t.Log("stdout", mockOut.String())
					if test.stdoutShouldContain != "" {
						require.Contains(t, mockOut.String(), test.stdoutShouldContain, "stdout should contain")
					}
					if test.stdoutShouldBe != "" {
						require.Equal(t, mockOut.String(), test.stdoutShouldBe, "stdout should be")
					}
				}

				if stderrShouldBeEmpty {
					require.Empty(t, mockErr.String(), "stderr should be empty")
				} else {
					t.Log("stderr", mockErr.String())
					if test.stderrShouldContain != "" {
						require.Contains(t, mockErr.String(), test.stderrShouldContain, "stderr should contain")
					}
					if test.stderrShouldBe != "" {
						require.Equal(t, mockErr.String(), test.stderrShouldBe, "stderr should be")
					}
				}
			}

			exec := "gnodev"
			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					t.Log("recover", output)
					require.False(t, recoverShouldBeEmpty, "should panic")
					require.True(t, errShouldBeEmpty, "should not return an error")
					if test.recoverShouldContain != "" {
						require.Contains(t, output, test.recoverShouldContain, "recover should contain")
					}
					if test.recoverShouldBe != "" {
						require.Equal(t, output, test.recoverShouldBe, "recover should be")
					}
					checkOutputs(t)
				} else {
					require.True(t, recoverShouldBeEmpty, "should not panic")
				}
			}()

			if test.simulateExternalRepo {
				// create external dir
				tmpDir, cleanUpFn := createTmpDir(t)
				defer cleanUpFn()

				// copy to external dir
				absTestDir, err := filepath.Abs(test.testDir)
				require.Nil(t, err)
				require.Nil(t, copyDir(absTestDir, tmpDir))

				// cd to tmp directory
				os.Chdir(tmpDir)
				defer os.Chdir(workingDir)
			}

			err := runMain(cmd, exec, test.args)

			if errShouldBeEmpty {
				require.Nil(t, err, "err should be nil")
			} else {
				t.Log("err", err.Error())
				require.NotNil(t, err, "err shouldn't be nil")
				if test.errShouldContain != "" {
					require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
				}
				if test.errShouldBe != "" {
					require.Equal(t, err.Error(), test.errShouldBe, "err should be")
				}
			}

			checkOutputs(t)
		})
	}
}
