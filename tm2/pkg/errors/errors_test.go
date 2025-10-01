package errors

import (
	"errors"
	fmt "fmt"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorPanic(t *testing.T) {
	t.Parallel()

	type pnk struct {
		msg string
	}

	capturePanic := func() (err Error) {
		defer func() {
			if r := recover(); r != nil {
				err = Wrap(r, "This is the message in Wrap(r, message).")
			}
		}()
		panic(pnk{"something"})
	}

	err := capturePanic()

	assert.Equal(t, pnk{"something"}, err.Data())
	assert.Equal(t, "{something}", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), "This is the message in Wrap(r, message).")
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestWrapSomething(t *testing.T) {
	t.Parallel()

	err := Wrapf("something", "formatter%v%v", 0, 1)

	assert.Equal(t, "something", err.Data())
	assert.Equal(t, "something", fmt.Sprintf("%v", err))
	assert.Regexp(t, `formatter01\n`, fmt.Sprintf("%#v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestWrapNothing(t *testing.T) {
	t.Parallel()

	err := Wrapf(nil, "formatter%v%v", 0, 1)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestErrorNew(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.NotContains(t, fmt.Sprintf("%#v", err), "Stack Trace")
}

func TestErrorNewWithDetails(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)
	err.Trace(0, "trace %v", 1)
	err.Trace(0, "trace %v", 2)
	err.Trace(0, "trace %v", 3)
	assert.Contains(t, fmt.Sprintf("%+v", err), `Data: formatter01`)
	assert.Contains(t, fmt.Sprintf("%+v", err), "Msg Traces:\n    0")
}

func TestErrorNewWithStacktrace(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1).Stacktrace()

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	assert.Contains(t, fmt.Sprintf("%#v", err), "Stack Trace:\n    0")
}

func TestErrorNewWithTrace(t *testing.T) {
	t.Parallel()

	err := New("formatter%v%v", 0, 1)
	err.Trace(0, "trace %v", 1)
	err.Trace(0, "trace %v", 2)
	err.Trace(0, "trace %v", 3)

	assert.Equal(t,
		FmtError{"formatter%v%v", []any{0, 1}},
		err.Data())
	assert.Equal(t, "formatter01", fmt.Sprintf("%v", err))
	assert.Contains(t, fmt.Sprintf("%#v", err), `Data: errors.FmtError{format:"formatter%v%v", args:[]interface {}{0, 1}}`)
	dump := fmt.Sprintf("%#v", err)
	assert.NotContains(t, dump, "Stack Trace")
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 1`, dump)
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 2`, dump)
	assert.Regexp(t, `errors/errors_test\.go:[0-9]+ - trace 3`, dump)
}

func TestWrapError(t *testing.T) {
	t.Parallel()

	var err1 error = New("my message")
	var err2 error = Wrap(err1, "another message")
	assert.Equal(t, err1, err2)
	assert.True(t, errors.Is(err2, err1))

	err1 = fmt.Errorf("my message")
	err2 = Wrap(err1, "another message")
	assert.NotEqual(t, err1, err2)
	assert.True(t, errors.Is(err2, err1))
}

func TestStripBuildDir(t *testing.T) {
	// Save original env and restore after test
	origGOMOD := os.Getenv("GOMOD")
	defer func() {
		os.Setenv("GOMOD", origGOMOD)
		// Reset the build dir cache
		buildDir = ""
		buildDirOnce = sync.Once{}
	}()

	tests := []struct {
		name     string
		setupEnv func()
		input    string
		expected string
	}{
		// Project paths
		{
			name: "project file with GOMOD set",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/go/src/github.com/gnolang/gno/go.mod")
			},
			input:    "/home/user/go/src/github.com/gnolang/gno/tm2/pkg/errors/errors.go",
			expected: "gno/tm2/pkg/errors/errors.go",
		},
		{
			name: "project file in subdirectory",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/go/src/github.com/gnolang/gno/go.mod")
			},
			input:    "/home/user/go/src/github.com/gnolang/gno/gno.land/pkg/gnoland/app.go",
			expected: "gno/gno.land/pkg/gnoland/app.go",
		},
		{
			name: "project file when GOMOD is /dev/null",
			setupEnv: func() {
				os.Setenv("GOMOD", "/dev/null")
			},
			input:    "/home/user/go/src/github.com/gnolang/gno/tm2/pkg/errors/errors.go",
			expected: "/home/user/go/src/github.com/gnolang/gno/tm2/pkg/errors/errors.go",
		},

		// Go module paths
		{
			name: "go module dependency",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "/home/user/go/pkg/mod/golang.org/x/sync@v0.13.0/errgroup/errgroup.go",
			expected: "mod/golang.org/x/sync@v0.13.0/errgroup/errgroup.go",
		},
		{
			name: "go module with complex version",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "/usr/local/go/pkg/mod/github.com/stretchr/testify@v1.8.4/assert/assertions.go",
			expected: "mod/github.com/stretchr/testify@v1.8.4/assert/assertions.go",
		},
		{
			name: "go toolchain path",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "/home/user/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.1.linux-amd64/src/runtime/asm_amd64.s",
			expected: "toolchain/runtime/asm_amd64.s",
		},
		{
			name: "go toolchain internal path",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "/usr/local/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.24.1.linux-amd64/src/runtime/proc.go",
			expected: "toolchain/runtime/proc.go",
		},

		// Edge cases
		{
			name: "path without any known prefix",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "/usr/lib/system/file.go",
			expected: "/usr/lib/system/file.go",
		},
		{
			name: "empty path",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "",
			expected: "",
		},
		{
			name: "relative path",
			setupEnv: func() {
				os.Setenv("GOMOD", "/home/user/project/go.mod")
			},
			input:    "relative/path/file.go",
			expected: "relative/path/file.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset the build dir cache for each test
			buildDir = ""
			buildDirOnce = sync.Once{}

			// Setup environment
			tt.setupEnv()

			// Test
			got := stripBuildDir(tt.input)
			if got != tt.expected {
				t.Errorf("stripBuildDir(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
