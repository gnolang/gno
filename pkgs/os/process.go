package os

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type Process struct {
	Label     string
	WorkDir   string
	ExecPath  string
	Args      []string
	Pid       int
	StartTime time.Time
	EndTime   time.Time
	Cmd       *exec.Cmd        `json:"-"`
	ExitState *os.ProcessState `json:"-"`
	Stdin     io.Reader        `json:"-"`
	Stdout    io.WriteCloser   `json:"-"`
	Stderr    io.WriteCloser   `json:"-"`
	WaitCh    chan struct{}    `json:"-"`
}

// execPath: command name
// args: args to command. (should not include name)
func StartProcess(label string, dir string, execPath string, args []string, stdin io.Reader, stdout, stderr io.WriteCloser) (*Process, error) {
	cmd := exec.Command(execPath, args...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Stdin = stdin
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	proc := &Process{
		Label:     label,
		WorkDir:   dir,
		ExecPath:  execPath,
		Args:      args,
		Pid:       cmd.Process.Pid,
		StartTime: time.Now(),
		Cmd:       cmd,
		ExitState: nil,
		Stdin:     stdin,
		Stdout:    stdout,
		Stderr:    stderr,
		WaitCh:    make(chan struct{}),
	}
	go func() {
		err := proc.Cmd.Wait()
		if err != nil {
			// fmt.Printf("Process exit: %v\n", err)
			if exitError, ok := err.(*exec.ExitError); ok {
				proc.ExitState = exitError.ProcessState
			}
		}
		proc.ExitState = proc.Cmd.ProcessState
		proc.EndTime = time.Now() // TODO make this goroutine-safe
		err = proc.Stdout.Close()
		if err != nil {
			fmt.Printf("Error closing stdout for %v: %v\n", proc.Label, err)
		}
		if proc.Stderr != proc.Stdout {
			err = proc.Stderr.Close()
			if err != nil {
				fmt.Printf("Error closing stderr for %v: %v\n", proc.Label, err)
			}
		}
		close(proc.WaitCh)
	}()
	return proc, nil
}

func (proc *Process) StopProcess(kill bool) error {
	defer func() {
		proc.Stdout.Close()
		if proc.Stderr != proc.Stdout {
			proc.Stderr.Close()
		}
	}()
	if kill {
		// fmt.Printf("Killing process %v\n", proc.Cmd.Process)
		return proc.Cmd.Process.Kill()
	} else {
		// fmt.Printf("Stopping process %v\n", proc.Cmd.Process)
		return proc.Cmd.Process.Signal(os.Interrupt)
	}
}
