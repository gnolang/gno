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

	choices := []Choice{
		{Key: "r", Aliases: []string{"realm"}, Description: "realm", IsDefault: false},
		{Key: "p", Aliases: []string{"package"}, Description: "package", IsDefault: true},
		{Key: "m", Aliases: []string{"main", "run"}, Description: "run script", IsDefault: false},
	}

	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"key r", "r\n", 0},
		{"alias realm", "realm\n", 0},
		{"key p", "p\n", 1},
		{"alias package", "package\n", 1},
		{"key m", "m\n", 2},
		{"alias main", "main\n", 2},
		{"alias run", "run\n", 2},
		{"empty selects default", "\n", 1},
		{"case insensitive", "R\n", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			io, _ := newMockIO(tt.input)
			got, err := PromptChoice(io, "Pick: ", choices)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("invalid retries then EOF", func(t *testing.T) {
		t.Parallel()
		io, stderr := newMockIO("xyz\n")
		_, err := PromptChoice(io, "Pick: ", choices)
		require.Error(t, err)
		assert.Contains(t, stderr.String(), `invalid choice: "xyz"`)
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
		assert.Equal(t, 0, got)
	})

	t.Run("empty returns no items error", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("")
		_, err := PromptSelect(io, "Template:", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no items")
	})

	t.Run("multi default", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, 0, got)
	})

	t.Run("multi by number", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("2\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, 1, got)
	})

	t.Run("multi by name", func(t *testing.T) {
		t.Parallel()
		io, _ := newMockIO("dao\n")
		got, err := PromptSelect(io, "Template:", multi)
		require.NoError(t, err)
		assert.Equal(t, 1, got)
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
