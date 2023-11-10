package commands

import (
	"errors"
	"os"
	"strings"

	"golang.org/x/term"
)

// GetPassword fetches the password using the provided prompt, if any
func (io *IOImpl) GetPassword(
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
func (io *IOImpl) readLine() (string, error) {
	input, err := io.inBuf.ReadString('\n')
	if err != nil {
		return "", err
	}

	return input[:len(input)-1], nil
}

// readPassword reads the password from a terminal
// without local echo
func readPassword() (string, error) {
	fd := int(os.Stdin.Fd())

	inputPass, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	return string(inputPass), nil
}

// GetConfirmation will request user give the confirmation from stdin.
// "y", "Y", "yes", "YES", and "Yes" all count as confirmations.
// If the input is not recognized, it returns false and a nil error.
func (io *IOImpl) GetConfirmation(prompt string) (bool, error) {
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

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func (io *IOImpl) GetCheckPassword(
	prompts [2]string,
	insecure bool,
) (string, error) {
	pass, err := io.GetPassword(prompts[0], insecure)
	if err != nil {
		return "", err
	}

	pass2, err := io.GetPassword(prompts[1], insecure)
	if err != nil {
		return "", err
	}

	if pass != pass2 {
		return "", errors.New("passphrases don't match")
	}

	return pass, nil
}

// GetString simply returns the trimmed string output of a given reader.
func (io *IOImpl) GetString(prompt string) (string, error) {
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
