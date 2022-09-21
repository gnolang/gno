package command

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// GetPassword will prompt for a password one-time (to sign a tx).
// Passwords may be blank; user must validate.
func (cmd *Command) GetPassword(prompt string, insecureUseStdin bool) (pass string, err error) {
	if prompt != "" {
		// On stderr so it isn't part of bash output.
		cmd.ErrPrintln(prompt)
	}

	// insecure stdin.
	if insecureUseStdin {
		return cmd.readLineFromInBuf()
	}

	// secure prompt.
	pass, err = cmd.readPasswordFromInBuf()
	if err != nil {
		return "", err
	}
	return pass, nil
}

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func (cmd *Command) GetCheckPassword(prompt, prompt2 string) (string, error) {
	pass, err := cmd.GetPassword(prompt, false)
	if err != nil {
		return "", err
	}
	pass2, err := cmd.GetPassword(prompt2, false)
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
	// On stderr so it isn't part of bash output.
	cmd.ErrPrintfln("%s [y/n]:", prompt)

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
		// On stderr so it isn't part of bash output.
		cmd.ErrPrintln(prompt)
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
	pass, err := cmd.InBuf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return pass[:len(pass)-1], nil
}

func (cmd *Command) readPasswordFromInBuf() (string, error) {
	var fd int
	var pass string
	if cmd.In == os.Stdin {
		fd = syscall.Stdin
		inputPass, err := term.ReadPassword(fd)
		if err != nil {
			return "", err
		}
		pass = string(inputPass)
	} else {
		s, err := cmd.InBuf.ReadString('\n')
		if err != nil {
			return "", err
		}
		pass = s[:len(s)-1]
	}

	return pass, nil
}

// Println prints a line terminated by a newline.
func (cmd *Command) Println(args ...interface{}) {
	fmt.Fprintln(cmd.OutBuf, args...)
	cmd.OutBuf.Flush()
}

// Printf prints a formatted string without trailing newline.
func (cmd *Command) Printf(format string, args ...interface{}) {
	fmt.Fprintf(cmd.OutBuf, format, args...)
	cmd.OutBuf.Flush()
}

// Printfln prints a formatted string terminated by a newline.
func (cmd *Command) Printfln(format string, args ...interface{}) {
	fmt.Fprintf(cmd.OutBuf, format+"\n", args...)
	cmd.OutBuf.Flush()
}

// ErrPrintln prints a line terminated by a newline to
// cmd.Err(Buf).
func (cmd *Command) ErrPrintln(args ...interface{}) {
	fmt.Fprintln(cmd.ErrBuf, args...)
	cmd.ErrBuf.Flush()
}

// ErrPrintfln prints a formatted string terminated by a newline to cmd.Err(Buf).
func (cmd *Command) ErrPrintfln(format string, args ...interface{}) {
	fmt.Fprintf(cmd.ErrBuf, format+"\n", args...)
	cmd.ErrBuf.Flush()
}
