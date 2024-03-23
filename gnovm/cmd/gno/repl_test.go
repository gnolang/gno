package main

import (
	"os/exec"
	"reflect"
	"testing"
)

func TestReplApp(t *testing.T) {
	tc := []testMainCase{
		{args: []string{"repl", "invalid-arg"}, errShouldBe: "flag: help requested"},

		// args
		// {args: []string{"repl", "..."}, stdoutShouldContain: "..."},
	}
	testMainCaseRun(t, tc)
}

func TestUpdateIndentLevel(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		indentLevel int
		want        int
	}{
		{
			name:        "Test with no brackets",
			line:        "Hello, World!",
			indentLevel: 0,
			want:        0,
		},
		{
			name:        "Test with open brackets",
			line:        "func main() {",
			indentLevel: 0,
			want:        1,
		},
		{
			name:        "Test with closed brackets",
			line:        "}",
			indentLevel: 1,
			want:        0,
		},
		{
			name:        "Test with colon",
			line:        "case 'a':",
			indentLevel: 0,
			want:        1,
		},
		{
			name:        "Test with multiple open brackets",
			line:        "func main() { if true {",
			indentLevel: 0,
			want:        2,
		},
		{
			name:        "Test with multiple closed brackets",
			line:        "} }",
			indentLevel: 2,
			want:        0,
		},
		{
			name:        "Test with mixed brackets",
			line:        "} else {",
			indentLevel: 1,
			want:        1,
		},
		{
			name:        "Test with no change in indent level",
			line:        "fmt.Println(\"Hello, World!\")",
			indentLevel: 1,
			want:        1,
		},
		{
			name:        "Test with colon and open bracket",
			line:        "case 'a': {",
			indentLevel: 0,
			want:        1,
		},
		{
			name:        "Test with colon and closed bracket",
			line:        "case 'a': }",
			indentLevel: 1,
			want:        0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateIndentLevel(tt.line, tt.indentLevel); got != tt.want {
				t.Errorf("updateIndentLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockCommandExecutor struct {
	ExecutedCommand *exec.Cmd
}

func (e *MockCommandExecutor) Execute(cmd *exec.Cmd) error {
	e.ExecutedCommand = cmd
	return nil
}

type MockOSGetter struct {
	OS string
}

func (m MockOSGetter) Get() string {
	return m.OS
}

func TestClearScreen(t *testing.T) {
	tests := []struct {
		name     string
		osGetter OsGetter
		expected []string
	}{
		{"Windows", MockOSGetter{OS: "windows"}, []string{"cmd", "/c", "cls"}},
		{"Other", MockOSGetter{OS: "linux"}, []string{"clear"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock executor
			executor := &MockCommandExecutor{}

			// Call the function under test with the mock OS getter
			clearScreen(executor, tt.osGetter)

			// Check that the correct command was executed
			if executor.ExecutedCommand == nil {
				t.Fatal("Expected a command to be executed, but it was not")
			}
			if !reflect.DeepEqual(executor.ExecutedCommand.Args, tt.expected) {
				t.Errorf("Expected command %v, but got %v", tt.expected, executor.ExecutedCommand.Args)
			}
		})
	}
}
