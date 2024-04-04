package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"reflect"
	"syscall"
	"testing"
	"unsafe"

	"github.com/creack/pty"
)

func TestReplApp(t *testing.T) {
	t.Parallel()
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

// captureStdout captures the output written to stdout
func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPutChar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    byte
		expected string
	}{
		{'A', "A"},
		{'B', "B"},
		{'1', "1"},
		{'\n', "\n"},
	}

	for _, test := range tests {
		output := captureStdout(func() {
			err := putChar(test.input)
			if err != nil {
				t.Errorf("putChar(%q) returned an error: %v", test.input, err)
			}
		})
		if output != test.expected {
			t.Errorf("putChar(%q) = %q, want %q", test.input, output, test.expected)
		}
	}
}

func TestPutChars(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    []byte
		expected string
	}{
		{[]byte("Hello"), "Hello"},
		{[]byte("World"), "World"},
		{[]byte("12345"), "12345"},
		{[]byte("\nNewLine"), "\nNewLine"},
	}

	for _, test := range tests {
		output := captureStdout(func() {
			err := putChars(test.input)
			if err != nil {
				t.Errorf("putChars(%q) returned an error: %v", string(test.input), err)
			}
		})
		if output != test.expected {
			t.Errorf("putChars(%q) = %q, want %q", string(test.input), output, test.expected)
		}
	}
}

func TestPeekChar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		prepFunc func()
		expected byte
		ok       bool
	}{
		{
			name: "available byte",
			prepFunc: func() {
				input = make(chan byte, 1)
				input <- 'A'
			},
			expected: 'A',
			ok:       true,
		},
		{
			name: "no available byte - timeout",
			prepFunc: func() {
				input = make(chan byte)
			},
			expected: 0,
			ok:       false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.prepFunc()

			b, ok := peekChar()

			if b != tc.expected || ok != tc.ok {
				t.Errorf("peekChar() = (%q, %v), want (%q, %v)", b, ok, tc.expected, tc.ok)
			}

			lastInOk = false
		})
	}
}

func TestGetChar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		setupFunc func()
		expected  byte
	}{
		{
			name: "get from lastIn",
			setupFunc: func() {
				lastIn = 'A'
				lastInOk = true
			},
			expected: 'A',
		},
		{
			name: "get from channel",
			setupFunc: func() {
				input = make(chan byte, 1)
				input <- 'B'
				lastInOk = false
			},
			expected: 'B',
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupFunc()
			result := getChar()

			if result != tc.expected {
				t.Errorf("getChar() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestMakeRaw(t *testing.T) {
	pty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("Failed to open pty: %v", err)
	}
	defer pty.Close()
	defer tty.Close()

	fd := int(tty.Fd())

	savedState, err := makeRaw(fd)
	if err != nil {
		t.Fatalf("MakeRaw failed: %v", err)
	}
	defer func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(setTermios), uintptr(unsafe.Pointer(&savedState.termios)), 0, 0, 0)
	}()

	// Check if the terminal is in raw mode
	var newState terminalState
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&newState.termios)), 0, 0, 0); err != 0 {
		t.Fatalf("Failed to get terminal settings: %v", err)
	}
	if newState.termios.Iflag&(syscall.ISTRIP|syscall.INLCR|syscall.ICRNL|syscall.IGNCR|syscall.IXON|syscall.IXOFF) != 0 {
		t.Errorf("Input flags are not disabled")
	}
	if newState.termios.Lflag&(syscall.ECHO|syscall.ICANON|syscall.ISIG) != 0 {
		t.Errorf("Local flags are not disabled")
	}

	// Restore the terminal settings and check if they match the original settings
	syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(setTermios), uintptr(unsafe.Pointer(&savedState.termios)), 0, 0, 0)
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&newState.termios)), 0, 0, 0); err != 0 {
		t.Fatalf("Failed to get terminal settings: %v", err)
	}
	if newState.termios != savedState.termios {
		t.Errorf("Terminal settings are not restored correctly")
	}
}

func TestMakeCbreak(t *testing.T) {
	pty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("Failed to open pty: %v", err)
	}
	defer pty.Close()
	defer tty.Close()

	fd := int(tty.Fd())

	savedState, err := makeCbreak(fd)
	if err != nil {
		t.Fatalf("MakeCbreak failed: %v", err)
	}
	defer func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(setTermios), uintptr(unsafe.Pointer(&savedState.termios)), 0, 0, 0)
	}()

	// Check if the terminal is in cbreak mode
	var newState terminalState
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&newState.termios)), 0, 0, 0); err != 0 {
		t.Fatalf("Failed to get terminal settings: %v", err)
	}
	if newState.termios.Iflag&(syscall.ISTRIP|syscall.INLCR|syscall.ICRNL|syscall.IGNCR|syscall.IXON|syscall.IXOFF) != 0 {
		t.Errorf("Input flags are not disabled")
	}
	if newState.termios.Lflag&(syscall.ECHO|syscall.ICANON) != 0 {
		t.Errorf("Local flags are not disabled")
	}

	// Restore the terminal settings and check if they match the original settings
	syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(setTermios), uintptr(unsafe.Pointer(&savedState.termios)), 0, 0, 0)
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&newState.termios)), 0, 0, 0); err != 0 {
		t.Fatalf("Failed to get terminal settings: %v", err)
	}
	if newState.termios != savedState.termios {
		t.Errorf("Terminal settings are not restored correctly")
	}
}

func TestRestore(t *testing.T) {
    pty, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("Failed to open pty: %v", err)
	}
	defer pty.Close()
	defer tty.Close()

	fd := int(tty.Fd())

    // Save the current terminal state
    var originalState terminalState
    if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&originalState.termios)), 0, 0, 0); err != 0 {
        t.Fatalf("Failed to get terminal settings: %v", err)
    }

    // Modify the terminal state
    var modifiedState terminalState
    modifiedState.termios = originalState.termios
    modifiedState.termios.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
    modifiedState.termios.Lflag &^= syscall.ECHO | syscall.ICANON
    if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(setTermios), uintptr(unsafe.Pointer(&modifiedState.termios)), 0, 0, 0); err != 0 {
        t.Fatalf("Failed to set terminal settings: %v", err)
    }

    // Restore the terminal state
    err = restore(fd, &originalState)
    if err != nil {
        t.Fatalf("restore failed: %v", err)
    }

    // Check if the terminal state is restored correctly
    var restoredState terminalState
    if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(getTermios), uintptr(unsafe.Pointer(&restoredState.termios)), 0, 0, 0); err != 0 {
        t.Fatalf("Failed to get terminal settings: %v", err)
    }
    if restoredState.termios != originalState.termios {
        t.Errorf("Terminal settings are not restored correctly")
    }
}