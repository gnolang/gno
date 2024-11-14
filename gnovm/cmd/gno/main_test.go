package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/gnovm/pkg/gnopkgfetch"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/stretchr/testify/require"
)

func TestMain_Gno(t *testing.T) {
	tc := []testMainCase{
		{args: []string{""}, errShouldBe: "flag: help requested"},
	}

	testMainCaseRun(t, tc)
}

type testMainCase struct {
	args                 []string
	testDir              string
	simulateExternalRepo bool
	noTmpGnohome         bool

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

	oldClient := gnopkgfetch.Client
	gnopkgfetch.Client = client.NewRPCClient(&examplesMockClient{})
	t.Cleanup(func() {
		gnopkgfetch.Client = oldClient
	})

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
			mockOut := bytes.NewBufferString("")
			mockErr := bytes.NewBufferString("")

			if !test.noTmpGnohome {
				tmpGnoHome, err := os.MkdirTemp(os.TempDir(), "gnotesthome_")
				require.NoError(t, err)
				t.Cleanup(func() { os.RemoveAll(tmpGnoHome) })
				t.Setenv("GNOHOME", tmpGnoHome)
			}

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
						require.Equal(t, test.stdoutShouldBe, mockOut.String(), "stdout should be")
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
						require.Equal(t, test.stderrShouldBe, mockErr.String(), "stderr should be")
					}
				}
			}

			defer func() {
				if r := recover(); r != nil {
					output := fmt.Sprintf("%v", r)
					t.Log("recover", output)
					require.False(t, recoverShouldBeEmpty, "should not panic")
					require.True(t, errShouldBeEmpty, "should not return an error")
					if test.recoverShouldContain != "" {
						require.Regexpf(t, test.recoverShouldContain, output, "recover should contain")
					}
					if test.recoverShouldBe != "" {
						require.Equal(t, test.recoverShouldBe, output, "recover should be")
					}
					checkOutputs(t)
				} else {
					require.True(t, recoverShouldBeEmpty, "should panic")
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

			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			err := newGnocliCmd(io).ParseAndRun(context.Background(), test.args)

			if errShouldBeEmpty {
				require.Nil(t, err, "err should be nil")
			} else {
				t.Log("err", fmt.Sprintf("%v", err))
				require.NotNil(t, err, "err shouldn't be nil")
				if test.errShouldContain != "" {
					require.Contains(t, err.Error(), test.errShouldContain, "err should contain")
				}
				if test.errShouldBe != "" {
					require.Equal(t, test.errShouldBe, err.Error(), "err should be")
				}
			}

			checkOutputs(t)
		})
	}
}

type examplesMockClient struct{}

func (m *examplesMockClient) SendRequest(ctx context.Context, request types.RPCRequest) (*types.RPCResponse, error) {
	params := struct {
		Path string `json:"path"`
		Data []byte `json:"data"`
	}{}
	if err := json.Unmarshal(request.Params, &params); err != nil {
		return nil, fmt.Errorf("failed to unmarshal params: %w", err)
	}
	path := params.Path
	if path != "vm/qfile" {
		return nil, fmt.Errorf("unexpected call to %q", path)
	}
	data := string(params.Data)

	examplesDir := filepath.Join(gnoenv.RootDir(), "examples")
	target := filepath.Join(examplesDir, data)

	res := ctypes.ResultABCIQuery{}

	finfo, err := os.Stat(target)
	if os.IsNotExist(err) {
		res.Response = sdk.ABCIResponseQueryFromError(fmt.Errorf("package %q is not available", data))
		return &types.RPCResponse{
			Result: amino.MustMarshalJSON(res),
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat %q: %w", data, err)
	}

	if finfo.IsDir() {
		entries, err := os.ReadDir(target)
		if err != nil {
			return nil, fmt.Errorf("failed to get package %q: %w", data, err)
		}
		files := []string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				files = append(files, entry.Name())
			}
		}
		res.Response.Data = []byte(strings.Join(files, "\n"))
	} else {
		content, err := os.ReadFile(target)
		if err != nil {
			return nil, fmt.Errorf("failed to get file %q: %w", data, err)
		}
		res.Response.Data = content
	}

	return &types.RPCResponse{
		Result: amino.MustMarshalJSON(res),
	}, nil
}

func (m *examplesMockClient) SendBatch(ctx context.Context, requests types.RPCRequests) (types.RPCResponses, error) {
	return nil, errors.New("not implemented")
}

func (m *examplesMockClient) Close() error {
	return nil
}
