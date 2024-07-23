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
	t.Parallel()
	tests := []struct {
		name                string
		line                string
		startingIndentLevel int
		want                int
	}{
		{
			name: "Test with no brackets",
			line: "Hello, World!",
		},
		{
			name: "Test with open brackets",
			line: "func main() {",
			want: 1,
		},
		{
			name:                "Test with closed brackets",
			line:                "}",
			startingIndentLevel: 1,
		},
		{
			name: "Test with colon",
			line: "case 'a':",
		},
		{
			name:                "Test with colon and closed bracket",
			line:                "case 'a': }",
			startingIndentLevel: 1,
		},
		{
			name: "Test with colon in string",
			line: "\"case 'a':\"",
		},
		{
			name: "Test with colon in string and string end with colon",
			line: "case ':':",
		},
		{
			name: "Test with multiple open brackets",
			line: "func main() { if true {",
			want: 2,
		},
		{
			name:                "Test with multiple closed brackets",
			line:                "} }",
			startingIndentLevel: 2,
		},
		{
			name:                "Test with mixed brackets",
			line:                "} else {",
			startingIndentLevel: 1,
			want:                1,
		},
		{
			name:                "Test with no change in indent level",
			line:                "fmt.Println(\"Hello, World!\")",
			startingIndentLevel: 1,
			want:                1,
		},
		{
			name: "Test with colon and open bracket",
			line: "case 'a': {",
			want: 1,
		},
		{
			name: "Test with brackets in string",
			line: "\"}}}}\"",
		},
		{
			name: "Test with brackets in single line comment",
			line: "// { [ (",
		},
		{
			name: "Test with brackets in multi line comment",
			line: "/* {{{{ */",
		},
		{
			name: "Test with brackets in string and comment",
			line: "ufmt.Println(\"{ [ ( ) ] } {{\") // { [ ( ) ] ",
		},
		{
			name: "Test string and single line comment",
			line: "CurlyToken = '{' // {",
		},
		{
			name: "Test curly bracket in string",
			line: "a := '{'",
		},
		{
			name: "Test curly bracket in string 2",
			line: "a := \"{hello\"",
		},
		{
			name: "Test with backticks",
			line: "`Hello, World!`",
		},
		{
			name: "Test with brackets in backticks",
			line: "`{([`",
		},
		{
			name: "Test with escaped brackets in backticks",
			line: "`\\{\\(\\[`",
		},
		{
			name: "Test with backticks and brackets",
			line: "func main() { `{([` }",
		},
		{
			name: "Test with escaped brackets",
			line: "c := \"asdf\\\\\"{{{{\"",
			want: 4,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := updateIndentLevel(tt.line, tt.startingIndentLevel); got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
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
	t.Parallel()
	tests := []struct {
		name     string
		osGetter OsGetter
		expected []string
	}{
		{"Windows", MockOSGetter{OS: "windows"}, []string{"cmd", "/c", "cls"}},
		{"Other", MockOSGetter{OS: "linux"}, []string{"clear"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
