package command

import (
	"errors"
	"fmt"
	"strings"
)

// MinPassLength is the minimum acceptable password length
const MinPassLength = 8

// GetPassword will prompt for a password one-time (to sign a tx)
// It enforces the password length
func (cmd *Command) GetPassword(prompt string) (pass string, err error) {
	pass, err = cmd.readLineFromInBuf()

	if err != nil {
		return "", err
	}

	if len(pass) < MinPassLength {
		// Return the given password to the upstream client so it can handle a
		// non-STDIN failure gracefully.
		return pass, fmt.Errorf("password must be at least %d characters", MinPassLength)
	}

	return pass, nil
}

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func (cmd *Command) GetCheckPassword(prompt, prompt2 string) (string, error) {
	pass, err := cmd.GetPassword(prompt)
	if err != nil {
		return "", err
	}
	pass2, err := cmd.GetPassword(prompt2)
	if err != nil {
		return "", err
	}
	if pass != pass2 {
		return "", errors.New("passphrases don't match")
	}
	return pass, nil
}

// GetConfirmation will request user give the confirmation from stdin.
// "y", "Y", "yes", "YES", and "Yes" all count as confirmations.
// If the input is not recognized, it returns false and a nil error.
func (cmd *Command) GetConfirmation(prompt string) (bool, error) {
	cmd.OutBuf.WriteString(fmt.Sprintf("%s [y/N]: ", prompt))

	response, err := cmd.readLineFromInBuf()
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

// GetString simply returns the trimmed string output of a given reader.
func (cmd *Command) GetString(prompt string) (string, error) {
	if prompt != "" {
		cmd.PrintPrefixed(prompt)
	}

	out, err := cmd.readLineFromInBuf()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// readLineFromInBuf reads one line from stdin.
// Subsequent calls reuse the same buffer, so we don't lose
// any input when reading a password twice (to verify)
func (cmd *Command) readLineFromInBuf() (string, error) {
	buf := cmd.InBuf
	pass, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

// PrintPrefixed prints a string with > prefixed for use in prompts.
func (cmd *Command) PrintPrefixed(msg string) {
	msg = fmt.Sprintf("> %s\n", msg)
	fmt.Fprint(cmd.OutBuf, msg)
	cmd.OutBuf.Flush()
}

// Println prints a line terminated by a newline.
func (cmd *Command) Println(line string) {
	fmt.Fprint(cmd.OutBuf, line+"\n")
	cmd.OutBuf.Flush()
}

// Printfln prints a formatted string terminated by a newline.
func (cmd *Command) Printfln(format string, args ...interface{}) {
	fmt.Fprintf(cmd.OutBuf, format+"\n", args...)
	cmd.OutBuf.Flush()
}

// ErrPrintPrefixed prints a string with > prefixed for use in prompts to cmd.Err(Buf).
func (cmd *Command) ErrPrintPrefixed(msg string) {
	msg = fmt.Sprintf("> %s\n", msg)
	fmt.Fprint(cmd.ErrBuf, msg)
	cmd.ErrBuf.Flush()
}

// ErrPrintln prints a line terminated by a newline to
// cmd.Err(Buf).
func (cmd *Command) ErrPrintln(line string) {
	fmt.Fprint(cmd.ErrBuf, line+"\n")
	cmd.ErrBuf.Flush()
}

// ErrPrintfln prints a formatted string terminated by a newline to cmd.Err(Buf).
func (cmd *Command) ErrPrintfln(format string, args ...interface{}) {
	fmt.Fprintf(cmd.ErrBuf, format+"\n", args...)
	cmd.ErrBuf.Flush()
}
