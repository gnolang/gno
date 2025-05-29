package test_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

type pkgTestCase struct {
	name                string
	pkgPath             string // relative to repo root
	runFlag             string // optional
	verbose             bool
	stderrShouldContain []string
	stdoutShouldContain []string
	errShouldBe         string
}

func TestPkgTestCases(t *testing.T) {
	rootDir := gnoenv.RootDir()

	cases := []pkgTestCase{
		{
			name:                "unit tests with one pass and one fail",
			pkgPath:             "../../tests/integ/test/basic",
			verbose:             true,
			stderrShouldContain: []string{"PASS: TestBasic/greater_than_one__", "FAIL: TestBasic/less_than_one"},
			errShouldBe:         `failed: "TestBasic"`,
		},
		{
			name:                "filtered unit tests using -run",
			pkgPath:             "../../tests/integ/test/basic",
			runFlag:             "TestBasic/gr[e](at|])er\\_.*|nonExistantPattern",
			verbose:             true,
			stderrShouldContain: []string{"PASS: TestBasic/greater_than_one__"},
			errShouldBe:         "",
		},
		{
			name:    "filetests: one pass, one fail",
			pkgPath: "../../tests/integ/test/basic_ft",
			verbose: true,
			stderrShouldContain: []string{
				"--- PASS: ../../tests/integ/test/basic_ft/pass_filetest.gno",
				"--- FAIL: ../../tests/integ/test/basic_ft/fail_filetest.gno",
			},
			stdoutShouldContain: []string{"Goodbye from filetest"},
			errShouldBe:         "../../tests/integ/test/basic_ft/fail_filetest.gno failed",
		},
		{
			name:                "unit tests has bad import",
			pkgPath:             "../../tests/integ/test/has_bad_import",
			verbose:             true,
			errShouldBe:         "unknown import path \"gno.land/q/there_is_no_q\"",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockOut := new(bytes.Buffer)
			mockErr := new(bytes.Buffer)
			io := commands.NewTestIO()
			io.SetOut(commands.WriteNopCloser(mockOut))
			io.SetErr(commands.WriteNopCloser(mockErr))

			opts := test.NewTestOptions(rootDir, io.Out(), io.Err())
			opts.Verbose = tc.verbose
			opts.RunFlag = tc.runFlag

			subPkgs, err := gnomod.SubPkgsFromPaths([]string{tc.pkgPath})
			if err != nil {
				t.Fatalf("list sub packages error: %v", err)
			}
			if len(subPkgs) == 0 {
				t.Fatalf("expected at least one package, got 0")
			}

			for _, pkg := range subPkgs {
				if len(pkg.TestGnoFiles) == 0 && len(pkg.FiletestGnoFiles) == 0 {
					t.Fatalf("no test files found in %q", pkg.Dir)
				}
				modfile, _ := gnomod.ParseDir(pkg.Dir)
				if modfile == nil {
					t.Fatalf("unable to parse gno.mod at %q", pkg.Dir)
				}
				gnoPkgPath := modfile.Module.Mod.Path
				memPkg := gno.MustReadMemPackage(pkg.Dir, gnoPkgPath)

				err := test.Test(memPkg, pkg.Dir, opts)

				if tc.errShouldBe != "" {
					if err == nil {
						t.Errorf("expected error: %q, got nil", tc.errShouldBe)
					} else if !strings.Contains(err.Error(), tc.errShouldBe) {
						t.Errorf("error mismatch:\nwant substring: %q\ngot: %q", tc.errShouldBe, err.Error())
					}
				} else if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				stdoutStr := mockOut.String()
				stderrStr := mockErr.String()

				for _, want := range tc.stdoutShouldContain {
					if !strings.Contains(stdoutStr, want) {
						t.Errorf("stdout missing expected substring:\nwant: %q\ngot: %q", want, stdoutStr)
					}
				}
				for _, want := range tc.stderrShouldContain {
					if !strings.Contains(stderrStr, want) {
						t.Errorf("stderr missing expected substring:\nwant: %q\ngot: %q", want, stderrStr)
					}
				}
			}
		})
	}
}
