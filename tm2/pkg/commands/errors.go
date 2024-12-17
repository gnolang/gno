package commands

import (
	"strconv"
)

// ExitCodeError is an error to terminate the program without printing any error,
// but passing in the given exit code to os.Exit.
//
// [Command.ParseAndRun] will return any ExitCodeError encountered, but
// [Command.Execute] will handle it and return an appropriate error message.
type ExitCodeError int

func (e ExitCodeError) Error() string {
	return "exit code: " + strconv.Itoa(int(e))
}
