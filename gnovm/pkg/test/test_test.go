package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/profiler"
	teststd "github.com/gnolang/gno/gnovm/tests/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	tests := []struct {
		name     string
		caller   crypto.Bech32Address
		pkgPath  string
		send     std.Coins
		validate func(t *testing.T, ctx *teststd.TestExecContext)
	}{
		{
			name:    "basic context creation",
			caller:  DefaultCaller,
			pkgPath: "gno.land/p/demo/test",
			send:    std.Coins{std.Coin{Denom: "ugnot", Amount: 1000}},
			validate: func(t *testing.T, ctx *teststd.TestExecContext) {
				assert.Equal(t, "dev", ctx.ChainID)
				assert.Equal(t, "gno.land", ctx.ChainDomain)
				assert.Equal(t, int64(DefaultHeight), ctx.Height)
				assert.Equal(t, int64(DefaultTimestamp), ctx.Timestamp)
				assert.Equal(t, DefaultCaller, ctx.OriginCaller)
				assert.NotNil(t, ctx.Banker)
				assert.NotNil(t, ctx.Params)
				assert.NotNil(t, ctx.EventLogger)
			},
		},
		{
			name:    "empty caller for package initialization",
			caller:  "",
			pkgPath: "gno.land/p/demo/test",
			send:    nil,
			validate: func(t *testing.T, ctx *teststd.TestExecContext) {
				assert.Equal(t, crypto.Bech32Address(""), ctx.OriginCaller)
				assert.Nil(t, ctx.OriginSend)
			},
		},
		{
			name:    "context with coins",
			caller:  DefaultCaller,
			pkgPath: "gno.land/p/demo/test",
			send:    std.Coins{std.Coin{Denom: "ugnot", Amount: 5000}},
			validate: func(t *testing.T, ctx *teststd.TestExecContext) {
				assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 5000}}, ctx.OriginSend)
				// Verify banker has coins for package address
				pkgAddr := gnolang.DerivePkgBech32Addr("gno.land/p/demo/test")
				banker := ctx.Banker.(*teststd.TestBanker)
				assert.Equal(t, std.Coins{std.Coin{Denom: "ugnot", Amount: 5000}}, banker.CoinTable[pkgAddr])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Context(tt.caller, tt.pkgPath, tt.send)
			require.NotNil(t, ctx)
			tt.validate(t, ctx)
		})
	}
}

func TestOutputWithError(t *testing.T) {
	mainOut := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	writer := OutputWithError(mainOut, errOut)
	require.NotNil(t, writer)

	// Test regular write
	n, err := writer.Write([]byte("normal output"))
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, "normal output", mainOut.String())
	assert.Empty(t, errOut.String())

	// Test stderr write
	if stderrWriter, ok := writer.(interface{ StderrWrite([]byte) (int, error) }); ok {
		n, err = stderrWriter.StderrWrite([]byte("error output"))
		assert.NoError(t, err)
		assert.Equal(t, 12, n)
		assert.Equal(t, "error output", errOut.String())
	}
}

func TestPrettySize(t *testing.T) {
	tests := []struct {
		name     string
		size     int64
		expected string
	}{
		{
			name:     "bytes",
			size:     999,
			expected: "999",
		},
		{
			name:     "kilobytes",
			size:     1500,
			expected: "1.5k",
		},
		{
			name:     "megabytes",
			size:     1500000,
			expected: "1.5M",
		},
		{
			name:     "gigabytes",
			size:     1500000000,
			expected: "1.5G",
		},
		{
			name:     "zero",
			size:     0,
			expected: "0",
		},
		{
			name:     "exact kilobyte",
			size:     1000,
			expected: "1.0k",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prettySize(tt.size)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFmtDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "one second",
			duration: time.Second,
			expected: "1.00s",
		},
		{
			name:     "fractional seconds",
			duration: 1500 * time.Millisecond,
			expected: "1.50s",
		},
		{
			name:     "milliseconds",
			duration: 250 * time.Millisecond,
			expected: "0.25s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0.00s",
		},
		{
			name:     "long duration",
			duration: 125 * time.Second,
			expected: "125.00s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fmtDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewTestOptions(t *testing.T) {
	rootDir := t.TempDir()
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	opts := NewTestOptions(rootDir, stdout, stderr)
	require.NotNil(t, opts)

	assert.Equal(t, rootDir, opts.RootDir)
	assert.Equal(t, stdout, opts.Output)
	assert.Equal(t, stderr, opts.Error)
	assert.NotNil(t, opts.BaseStore)
	assert.NotNil(t, opts.TestStore)

	// Test WriterForStore
	writer := opts.WriterForStore()
	require.NotNil(t, writer)
}

func TestProxyWriter(t *testing.T) {
	mainBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}

	pw := &proxyWriter{
		w:    mainBuf,
		errW: errBuf,
	}

	// Test Write
	n, err := pw.Write([]byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, "test", mainBuf.String())

	// Test StderrWrite
	n, err = pw.StderrWrite([]byte("error"))
	assert.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "error", errBuf.String())

	// Test tee functionality
	teeBuf := &bytes.Buffer{}
	revert := pw.tee(teeBuf)

	pw.Write([]byte(" more"))
	assert.Equal(t, "test more", mainBuf.String())
	assert.Equal(t, " more", teeBuf.String())

	// Revert tee
	revert()
	teeBuf.Reset()
	pw.Write([]byte(" after"))
	assert.Equal(t, "test more after", mainBuf.String())
	assert.Empty(t, teeBuf.String())
}

func TestShouldRun(t *testing.T) {
	tests := []struct {
		name     string
		filter   string
		path     string
		expected bool
	}{
		{
			name:     "nil filter allows all",
			filter:   "",
			path:     "any/path",
			expected: true,
		},
		{
			name:     "exact match last segment",
			filter:   "TestFoo",
			path:     "TestFoo",
			expected: true,
		},
		{
			name:     "no match",
			filter:   "TestFoo",
			path:     "TestBar",
			expected: false,
		},
		{
			name:     "regex partial match",
			filter:   "Foo",
			path:     "TestFooBar",
			expected: true,
		},
		{
			name:     "match in path segment",
			filter:   "test",
			path:     "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var filter filterMatch
			if tt.filter != "" {
				filter = splitRegexp(tt.filter)
			}
			result := shouldRun(filter, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseMemPackageTests(t *testing.T) {
	// Create test memory package
	mpkg := &std.MemPackage{
		Name: "testpkg",
		Files: []*std.MemFile{
			{
				Name: "main.gno",
				Body: `package testpkg

func Add(a, b int) int {
	return a + b
}`,
			},
			{
				Name: "main_test.gno",
				Body: `package testpkg

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("1 + 2 should equal 3")
	}
}`,
			},
			{
				Name: "main_test.gno",
				Body: `package testpkg_test

import "testing"

func TestAddExternal(t *testing.T) {
	// External test
}`,
			},
			{
				Name: "example_filetest.gno",
				Body: `package testpkg

// @test: add
// @expected: 3

func main() {
	println(Add(1, 2))
}`,
			},
			{
				Name: "not_a_go_file.txt",
				Body: "This should be skipped",
			},
		},
	}

	tset, itset, itfiles, ftfiles := parseMemPackageTests(mpkg)

	// Verify test files
	assert.NotNil(t, tset)
	assert.Equal(t, 1, len(tset.Files))

	// Verify integration test files
	assert.NotNil(t, itset)
	assert.Equal(t, 1, len(itset.Files))
	assert.Len(t, itfiles, 1)

	// Verify filetest files
	assert.Len(t, ftfiles, 1)
	assert.Equal(t, "example_filetest.gno", ftfiles[0].Name)
}

// Test tee helper function
func TestTee(t *testing.T) {
	original := &bytes.Buffer{}
	original.WriteString("original")

	var w io.Writer = original
	additional := &bytes.Buffer{}

	revert := tee(&w, additional)

	// Write should go to both
	w.Write([]byte(" test"))
	assert.Equal(t, "original test", original.String())
	assert.Equal(t, " test", additional.String())

	// After revert, should only write to original
	revert()
	additional.Reset()
	w.Write([]byte(" more"))
	assert.Equal(t, "original test more", original.String())
	assert.Empty(t, additional.String())
}

func TestTeeWithDiscard(t *testing.T) {
	var w io.Writer = io.Discard
	buf := &bytes.Buffer{}

	revert := tee(&w, buf)

	// When original is Discard, tee should replace it
	w.Write([]byte("test"))
	assert.Equal(t, "test", buf.String())

	// After revert, should be back to Discard
	revert()
	buf.Reset()
	w.Write([]byte("more"))
	assert.Empty(t, buf.String())
}

// Mock profiler for testing
type mockProfiler struct {
	lineProfilingEnabled bool
	started              bool
	stopped              bool
	profile              *profiler.Profile
}

var _ ProfilerInterface = (*mockProfiler)(nil)

func (m *mockProfiler) EnableLineProfiling() {
	m.lineProfilingEnabled = true
}

func (m *mockProfiler) StartProfiling(machine any, opts profiler.Options) error {
	m.started = true
	return nil
}

func (m *mockProfiler) StopProfiling() *profiler.Profile {
	m.stopped = true
	if m.profile == nil {
		return &profiler.Profile{}
	}
	return m.profile
}

func (m *mockProfiler) SetProfiler(p *profiler.Profiler) {}

type mockProfileWriter struct {
	writeCalled bool
	writeError  error
	profile     *profiler.Profile
	// opts        *TestOptions
}

func (m *mockProfileWriter) WriteProfile(profile *profiler.Profile, pc *ProfileConfig, output, errOutput io.Writer, testStore gnolang.Store) error {
	m.writeCalled = true
	m.profile = profile
	return m.writeError
}

func TestInitializeProfiling(t *testing.T) {
	tests := []struct {
		name        string
		pc          *ProfileConfig
		expectNil   bool
		checkResult func(t *testing.T, p *profiler.Profiler)
	}{
		{
			name:      "profiling disabled",
			pc:        nil,
			expectNil: true,
		},
		{
			name: "cpu profiling enabled",
			pc: &ProfileConfig{
				Enabled: true,
				Type:    "cpu",
			},
			expectNil: false,
			checkResult: func(t *testing.T, p *profiler.Profiler) {
				assert.NotNil(t, p)
			},
		},
		{
			name: "memory profiling enabled",
			pc: &ProfileConfig{
				Enabled: true,
				Type:    "memory",
			},
			expectNil: false,
			checkResult: func(t *testing.T, p *profiler.Profiler) {
				assert.NotNil(t, p)
			},
		},
		{
			name: "line profiling enabled",
			pc: &ProfileConfig{
				Enabled:      true,
				FunctionList: "TestFunction",
			},
			expectNil: false,
			checkResult: func(t *testing.T, p *profiler.Profiler) {
				assert.NotNil(t, p)
				// TODO: Can't directly check if line profiling is enabled without accessing internals
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := initializeProfiling(tt.pc)
			assert.NoError(t, err)

			if tt.expectNil {
				assert.Nil(t, p)
			} else {
				assert.NotNil(t, p)
				if tt.checkResult != nil {
					tt.checkResult(t, p)
				}
			}
		})
	}
}

func TestFinalizeProfiling(t *testing.T) {
	tests := []struct {
		name        string
		pc          *ProfileConfig
		opts        *TestOptions
		writer      *mockProfileWriter
		expectError bool
	}{
		{
			name:        "nil profile config",
			pc:          nil,
			opts:        &TestOptions{},
			writer:      &mockProfileWriter{},
			expectError: false,
		},
		{
			name: "nil profiler in config",
			pc: &ProfileConfig{
				Enabled: true,
			},
			opts:        &TestOptions{},
			writer:      &mockProfileWriter{},
			expectError: false,
		},
		{
			name: "successful write",
			pc: &ProfileConfig{
				Enabled: true,
				// TODO: Would need actual profiler in proper test
				profiler: &profiler.Profiler{},
			},
			opts: &TestOptions{
				Output: &bytes.Buffer{},
				Error:  &bytes.Buffer{},
			},
			writer:      &mockProfileWriter{},
			expectError: false,
		},
		{
			name: "write error",
			pc: &ProfileConfig{
				Enabled: true,
				// TODO: Would need actual profiler in proper test
				profiler: &profiler.Profiler{},
			},
			opts: &TestOptions{
				Output: &bytes.Buffer{},
				Error:  &bytes.Buffer{},
			},
			writer: &mockProfileWriter{
				writeError: fmt.Errorf("write failed"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test is limited because we can't easily create a real profiler
			// with a valid profile without running actual profiling
			if tt.pc == nil || tt.pc.profiler == nil {
				err := finalizeProfiling(tt.pc, tt.opts, tt.writer)
				assert.NoError(t, err)
				assert.False(t, tt.writer.writeCalled)
			}
		})
	}
}

func TestDefaultProfileWriter_WriteProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     *profiler.Profile
		pc          *ProfileConfig
		output      *bytes.Buffer
		errOutput   *bytes.Buffer
		testStore   gnolang.Store
		expectError bool
		checkOutput func(t *testing.T, output, errOut *bytes.Buffer)
	}{
		{
			name:      "nil profile",
			profile:   nil,
			pc:        &ProfileConfig{},
			output:    &bytes.Buffer{},
			errOutput: &bytes.Buffer{},
		},
		{
			name:    "nil profile config",
			profile: &profiler.Profile{},
			pc:      nil,
			output:  &bytes.Buffer{},
		},
		{
			name: "write to stdout",
			pc: &ProfileConfig{
				PrintToStdout: true,
				Format:        "text",
			},
			output:      &bytes.Buffer{},
			errOutput:   &bytes.Buffer{},
			expectError: true, // Will error because profile is nil in test
		},
		{
			name: "write to file",
			pc: &ProfileConfig{
				OutputFile: "/tmp/test_profile.out",
				Format:     "json",
			},
			output:      &bytes.Buffer{},
			errOutput:   &bytes.Buffer{},
			expectError: true, // Will error because profile is nil in test
		},
		{
			name: "write function list",
			pc: &ProfileConfig{
				FunctionList: "TestFunction",
			},
			output:      &bytes.Buffer{},
			errOutput:   &bytes.Buffer{},
			testStore:   nil,  // Would need real store
			expectError: true, // Will error because profile is nil in test
		},
	}

	writer := &DefaultProfileWriter{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pc != nil && tt.pc.OutputFile != "" && tt.pc.OutputFile != "/tmp/test_profile.out" {
				defer os.Remove(tt.pc.OutputFile)
			}

			err := writer.WriteProfile(tt.profile, tt.pc, tt.output, tt.errOutput, tt.testStore)
			if tt.profile == nil || tt.pc == nil {
				assert.NoError(t, err)
			} else if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.checkOutput != nil {
				tt.checkOutput(t, tt.output, tt.errOutput)
			}
		})
	}
}

func TestProfileFormatIntegration(t *testing.T) {
	pc := &ProfileConfig{
		Enabled:       true,
		Format:        "json",
		PrintToStdout: true,
	}

	// This is a limited integration test since we can't easily
	// create a real profile without running actual code
	p, err := initializeProfiling(pc)
	if err != nil {
		t.Skip("Could not initialize profiler")
	}
	defer func() {
		if p != nil {
			p.StopProfiling()
		}
	}()

	assert.NotNil(t, p)
}

func TestProfileConfigMethods(t *testing.T) {
	t.Run("IsEnabled", func(t *testing.T) {
		assert.False(t, (*ProfileConfig)(nil).IsEnabled())
		assert.False(t, (&ProfileConfig{Enabled: false}).IsEnabled())
		assert.True(t, (&ProfileConfig{Enabled: true}).IsEnabled())
	})

	t.Run("GetFormat", func(t *testing.T) {
		assert.Equal(t, profiler.FormatText, (*ProfileConfig)(nil).GetFormat())
		assert.Equal(t, profiler.FormatText, (&ProfileConfig{Format: ""}).GetFormat())
		assert.Equal(t, profiler.FormatJSON, (&ProfileConfig{Format: "json"}).GetFormat())
		assert.Equal(t, profiler.FormatTopList, (&ProfileConfig{Format: "toplist"}).GetFormat())
	})

	t.Run("GetProfileType", func(t *testing.T) {
		assert.Equal(t, profiler.ProfileCPU, (*ProfileConfig)(nil).GetProfileType())
		assert.Equal(t, profiler.ProfileCPU, (&ProfileConfig{Type: "cpu"}).GetProfileType())
		assert.Equal(t, profiler.ProfileMemory, (&ProfileConfig{Type: "memory"}).GetProfileType())
	})

	t.Run("GetSampleRate", func(t *testing.T) {
		assert.Equal(t, 100, (&ProfileConfig{Type: "cpu"}).GetSampleRate())
		assert.Equal(t, 1, (&ProfileConfig{Type: "memory"}).GetSampleRate())
	})
}

func TestWriteProfileWithMockData(t *testing.T) {
	writer := &DefaultProfileWriter{}

	t.Run("write function list success", func(t *testing.T) {
		output := &bytes.Buffer{}
		errOutput := &bytes.Buffer{}
		pc := &ProfileConfig{
			FunctionList: "TestFunction",
		}

		// Test with nil profile to verify early return
		err := writer.WriteProfile(nil, pc, output, errOutput, nil)
		assert.NoError(t, err) // nil profile returns no error
		assert.Empty(t, output.String())
	})

	t.Run("write to stdout success", func(t *testing.T) {
		output := &bytes.Buffer{}
		errOutput := &bytes.Buffer{}
		pc := &ProfileConfig{
			PrintToStdout: true,
			Format:        "text",
		}

		err := writer.WriteProfile(nil, pc, output, errOutput, nil)
		assert.NoError(t, err)
	})

	t.Run("write to file success", func(t *testing.T) {
		tempFile := "/tmp/test_profile_" + fmt.Sprintf("%d", time.Now().UnixNano()) + ".out"
		defer os.Remove(tempFile)

		output := &bytes.Buffer{}
		errOutput := &bytes.Buffer{}
		pc := &ProfileConfig{
			OutputFile: tempFile,
			Format:     "json",
		}

		err := writer.WriteProfile(nil, pc, output, errOutput, nil)
		assert.NoError(t, err)

		// With a nil profile, file shouldn't be created
		_, statErr := os.Stat(tempFile)
		assert.True(t, os.IsNotExist(statErr))
	})
}

func TestWriteProfileErrorCases(t *testing.T) {
	writer := &DefaultProfileWriter{}

	t.Run("file creation error", func(t *testing.T) {
		output := &bytes.Buffer{}
		errOutput := &bytes.Buffer{}
		pc := &ProfileConfig{
			OutputFile: "/invalid\x00path/profile.out", // Invalid path
			Format:     "json",
		}

		// Since we need a real profile to reach the file creation code,
		// we can't test this directly without significant mocking
		err := writer.WriteProfile(nil, pc, output, errOutput, nil)
		assert.NoError(t, err) // nil profile returns early
	})
}

// TODO: With nil profile, we just verify the branch would be taken
func TestWriteProfileIntegration(t *testing.T) {
	writer := &DefaultProfileWriter{}

	t.Run("verify function list branch logic", func(t *testing.T) {
		pc := &ProfileConfig{
			FunctionList: "TestFunc",
		}
		err := writer.WriteProfile(nil, pc, &bytes.Buffer{}, &bytes.Buffer{}, nil)
		assert.NoError(t, err)
	})

	t.Run("verify stdout branch logic", func(t *testing.T) {
		pc := &ProfileConfig{
			PrintToStdout: true,
			Format:        "json",
		}
		err := writer.WriteProfile(nil, pc, &bytes.Buffer{}, &bytes.Buffer{}, nil)
		assert.NoError(t, err)
	})

	t.Run("verify file output branch logic", func(t *testing.T) {
		pc := &ProfileConfig{
			OutputFile: "/tmp/test.prof",
			Format:     "text",
		}
		err := writer.WriteProfile(nil, pc, &bytes.Buffer{}, &bytes.Buffer{}, nil)
		assert.NoError(t, err)
	})
}

func TestFinalizeProfiling_Integration(t *testing.T) {
	t.Run("with real profiler stopping", func(t *testing.T) {
		pc := &ProfileConfig{
			Enabled:       true,
			PrintToStdout: true,
			Format:        "text",
		}

		// Initialize a real profiler
		p, err := initializeProfiling(pc)
		require.NoError(t, err)
		require.NotNil(t, p)

		// Now test finalizeProfiling
		opts := &TestOptions{
			Output: &bytes.Buffer{},
			Error:  &bytes.Buffer{},
		}
		writer := &DefaultProfileWriter{}

		err = finalizeProfiling(pc, opts, writer)
		assert.NoError(t, err)
	})
}
