package commands

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/term"
)

// GetPassword fetches the password using the provided prompt, if any
func GetPassword(prompt string, insecure bool) (string, error) {
	if prompt != "" {
		// Print out the prompt
		fmt.Println(prompt)
	}

	if insecure {
		return ReadLine()
	}

	return readPassword()
}

// ReadLine reads a new line from standard input
func ReadLine() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return input[:len(input)-1], nil
}

// readPassword reads the password from a terminal
// without local echo
func readPassword() (string, error) {
	fd := syscall.Stdin

	inputPass, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	return string(inputPass), nil
}
