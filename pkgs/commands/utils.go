package commands

import (
	"bufio"
	"errors"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// GetPassword fetches the password using the provided prompt, if any
func (io *IO) GetPassword(
	prompt string,
	insecure bool,
) (string, error) {
	if prompt != "" {
		// Print out the prompt
		// On stderr, so it isn't part of bash output
		io.ErrPrintln(prompt)
	}

	if insecure {
		return io.readLine()
	}

	return readPassword()
}

// readLine reads a new line from standard input
func (io *IO) readLine() (string, error) {
	input, err := io.inBuf.ReadString('\n')
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

// GetConfirmation will request user give the confirmation from stdin.
// "y", "Y", "yes", "YES", and "Yes" all count as confirmations.
// If the input is not recognized, it returns false and a nil error.
func (io *IO) GetConfirmation(prompt string) (bool, error) {
	// On stderr so it isn't part of bash output.
	io.ErrPrintfln("%s [y/n]:", prompt)

	response, err := io.readLine()
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(response)
	if len(response) == 0 {
		return false, nil
	}

	response = strings.ToLower(response)
	if response[0] == 'y' {
		return true, nil
	}

	return false, nil
}

type passwordReader interface {
	readPassword() (string, error)
}

type insecurePasswordReader struct {
	reader *bufio.Reader
}

func (ipr *insecurePasswordReader) readPassword() (string, error) {
	password, err := ipr.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return password, nil
}

type terminalPasswordReader struct {
}

func (tpr *terminalPasswordReader) readPassword() (string, error) {
	return readPassword()
}

func (io *IO) confirmPassword(prompt, prompt2 string, reader passwordReader) (string, error) {
	if prompt != "" {
		io.Println(prompt)
	}

	firstRead, err := reader.readPassword()
	if err != nil {
		return "", err
	}

	firstPassword := firstRead[:len(firstRead)-1]

	if prompt2 != "" {
		io.Println(prompt2)
	}

	secondRead, err := reader.readPassword()
	if err != nil {
		return "", err
	}

	secondPassword := secondRead[:len(secondRead)-1]

	if firstPassword != secondPassword {
		return "", errors.New("passphrases don't match")
	}

	return firstPassword, nil
}

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func (io *IO) GetCheckPassword(
	prompt,
	prompt2 string,
	insecure bool,
) (string, error) {
	if insecure {
		return io.confirmPassword(prompt, prompt2, &insecurePasswordReader{
			reader: io.inBuf,
		})
	}

	return io.confirmPassword(prompt, prompt2, &terminalPasswordReader{})
}

// GetString simply returns the trimmed string output of a given reader.
func (io *IO) GetString(prompt string) (string, error) {
	if prompt != "" {
		// On stderr so it isn't part of bash output.
		io.ErrPrintln(prompt)
	}

	out, err := io.readLine()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}
