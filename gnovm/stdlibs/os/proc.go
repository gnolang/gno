package os

// Args hold the command-line arguments, starting with the program name.
var Args []string

func init() {
	Args = runtime_args()
}

func runtime_args() []string // in package runtime
