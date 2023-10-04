package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// This is for testing purpose only. We don't need to or should not use IOpipe in normal flow.
// We redirect os.Stdin for mocking tests so that we don't need to pass commands.IO, which
// included an os.Stdin, to all the server command. It is not safe to expose os.Stdin in the blockchain node

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
