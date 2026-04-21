package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockIO(stdin string) (IO, *strings.Builder) {
	var stderrBuf strings.Builder
	mockIO := NewTestIO()
	mockIO.SetIn(strings.NewReader(stdin))
	mockIO.SetOut(WriteNopCloser(os.Stdout))
	mockIO.SetErr(WriteNopCloser(&stderrBuf))
	return mockIO, &stderrBuf
}

func TestPromptString(t *testing.T) {
	t.Parallel()

	t.Run("accepts input", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("hello\n")
		got, err := PromptString(io, "Enter value", "", nil)
		require.NoError(t, err)
		assert.Equal(t, "hello", got)
	})

	t.Run("returns default on empty", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("\n")
		got, err := PromptString(io, "Enter value", "mydefault", nil)
		require.NoError(t, err)
		assert.Equal(t, "mydefault", got)
	})

	t.Run("validates and retries", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("bad\ngood\n")
		validate := func(s string) error {
			if s == "bad" {
				return assert.AnError
			}
			return nil
		}
		got, err := PromptString(io, "Enter value", "", validate)
		require.NoError(t, err)
		assert.Equal(t, "good", got)
		assert.Contains(t, stderr.String(), assert.AnError.Error())
	})

	t.Run("validates empty rejected", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("\n")
		validate := func(s string) error {
			if s == "" {
				return assert.AnError
			}
			return nil
		}
		_, err := PromptString(io, "Enter value", "", validate)
		require.Error(t, err)
		assert.Contains(t, stderr.String(), assert.AnError.Error())
	})

	t.Run("EOF returns error", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("")
		_, err := PromptString(io, "Enter value", "", nil)
		require.Error(t, err)
	})
}

func TestPromptChoice(t *testing.T) {
	t.Parallel()

	choices := map[string]Choice{
		"r": {Aliases: []string{"realm"}, Description: "realm"},
		"p": {Aliases: []string{"package"}, Description: "package"},
		"m": {Aliases: []string{"main", "run"}, Description: "run script"},
	}

	tests := []struct {
		name       string
		input      string
		defaultKey string
		want       string
	}{
		{"key r", "r\n", "", "r"},
		{"alias realm", "realm\n", "", "r"},
		{"key p", "p\n", "", "p"},
		{"alias package", "package\n", "", "p"},
		{"key m", "m\n", "", "m"},
		{"alias main", "main\n", "", "m"},
		{"alias run", "run\n", "", "m"},
		{"empty selects default", "\n", "p", "p"},
		{"case insensitive", "R\n", "", "r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			io, _ := newMockIO(tt.input)
			got, err := PromptChoice(io, "Pick: ", choices, tt.defaultKey)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("invalid retries then EOF", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("xyz\n")
		_, err := PromptChoice(io, "Pick: ", choices, "")
		require.Error(t, err)
		assert.Contains(t, stderr.String(), `invalid choice: "xyz"`)
	})

	t.Run("empty no default", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("\n")
		_, err := PromptChoice(io, "Pick: ", choices, "")
		require.Error(t, err)
		assert.Contains(t, stderr.String(), "please enter a valid choice")
	})
}

func TestPromptSelect(t *testing.T) {
	t.Parallel()

	single := []SelectItem{
		{Name: "basic", Description: "test template"},
	}
	multi := []SelectItem{
		{Name: "basic", Description: "basic desc"},
		{Name: "dao", Description: "dao desc"},
	}

	t.Run("single auto-selects", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("")
		got, err := PromptSelect(io, "Template:", single)
		require.NoError(t, err)
		assert.Equal(t, "basic", got)
	})

	t.Run("empty returns no items error", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("")
		_, err := PromptSelect(io, "Template:", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no items")
	})

	t.Run("multi default first", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, "basic", got)
	})

	t.Run("multi by number", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("2\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, "dao", got)
	})

	t.Run("multi by name", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("dao\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, "dao", got)
	})

	t.Run("invalid number retries then EOF", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("99\n")
		_, err := PromptSelect(io, "Template:", multi)
		require.Error(t, err)
		assert.Contains(t, stderr.String(), "invalid choice: 99")
	})

	t.Run("unknown name retries then EOF", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("unknown\n")
		_, err := PromptSelect(io, "Template:", multi)
		require.Error(t, err)
		assert.Contains(t, stderr.String(), `unknown choice: "unknown"`)
	})
}
