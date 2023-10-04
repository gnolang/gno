package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// This is for testing purposes only.
// For mocking tests, we redirect os.Stdin so that we don't need to pass commands.IO,
// which includes os.Stdin, to all the server commands. Exposing os.Stdin in a blockchain node is not safe.
// This replaces the global variable and should not be used in concurrent tests. It's intended to simulate CLI input.
// We purposely avoid using a mutex to prevent giving the wrong impression that it's suitable for parallel tests.

type MockStdin struct {
	origStdout   *os.File
	stdoutReader *os.File

	outCh chan []byte

	origStdin   *os.File
	stdinWriter *os.File
}

func NewMockStdin(input string) (*MockStdin, error) {
	// Pipe for stdin. w ( stdinWriter ) -> r (stdin)
	stdinReader, stdinWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Pipe for stdout. w( stdout ) -> r (stdoutReader)
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	origStdin := os.Stdin
	os.Stdin = stdinReader

	_, err = stdinWriter.Write([]byte(input))
	if err != nil {
		stdinWriter.Close()
		os.Stdin = origStdin
		return nil, err
	}

	origStdout := os.Stdout
	os.Stdout = stdoutWriter

	outCh := make(chan []byte)

	// This goroutine reads stdout into a buffer in the background.
	go func() {
		var b bytes.Buffer
		if _, err := io.Copy(&b, stdoutReader); err != nil {
			log.Println(err)
		}
		outCh <- b.Bytes()
	}()

	return &MockStdin{
		origStdout:   origStdout,
		stdoutReader: stdoutReader,
		outCh:        outCh,
		origStdin:    origStdin,
		stdinWriter:  stdinWriter,
	}, nil
}

// ReadAndRestore collects all captured stdout and returns it; it also restores
// os.Stdin and os.Stdout to their original values.
func (i *MockStdin) ReadAndClose() ([]byte, error) {
	if i.stdoutReader == nil {
		return nil, fmt.Errorf("ReadAndRestore from closed FakeStdio")
	}

	// Close the writer side of the faked stdout pipe. This signals to the
	// background goroutine that it should exit.
	os.Stdout.Close()
	out := <-i.outCh

	os.Stdout = i.origStdout
	os.Stdin = i.origStdin

	if i.stdoutReader != nil {
		i.stdoutReader.Close()
		i.stdoutReader = nil
	}

	if i.stdinWriter != nil {
		i.stdinWriter.Close()
		i.stdinWriter = nil
	}

	return out, nil
}

// Call this in a defer function to restore and close os.Stdout and os.Stdin.
// This acts as a safeguard.
func (i *MockStdin) Close() {
	os.Stdout = i.origStdout
	os.Stdin = i.origStdin

	if i.stdoutReader != nil {
		i.stdoutReader.Close()
		i.stdoutReader = nil
	}

	if i.stdinWriter != nil {
		i.stdinWriter.Close()
		i.stdinWriter = nil
	}
}
