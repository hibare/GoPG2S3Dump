// Package exec provides an interface and implementation for executing system commands.
package exec

import (
	"context"
	"os"
	"os/exec"
)

// ExecIface defines the interface for building and running commands.
// revive:disable-next-line exported
type ExecIface interface {
	Command(ctx context.Context, name string, args ...string) CmdIface
	LookPath(file string) (string, error)
}

// CmdIface defines a single prepared command.
type CmdIface interface {
	WithEnv(env []string) CmdIface
	WithDir(dir string) CmdIface
	WithStdout(stdout *os.File) CmdIface
	WithStderr(stderr *os.File) CmdIface

	Run() error
	Output() ([]byte, error)
	CombinedOutput() ([]byte, error)
}

// Exec executes real system commands.
type Exec struct{}

// Command creates a new command with the given context, name, and arguments.
func (Exec) Command(ctx context.Context, name string, args ...string) CmdIface {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ() // default to system env
	return &Cmd{cmd: cmd}
}

// LookPath searches for an executable in the system's PATH.
func (Exec) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// Cmd wraps exec.Cmd with builder-style config.
type Cmd struct {
	cmd *exec.Cmd
}

// WithEnv sets environment variables for the command.
func (c *Cmd) WithEnv(env []string) CmdIface {
	c.cmd.Env = append(os.Environ(), env...) // merge system env + custom
	return c
}

// WithDir sets the working directory for the command.
func (c *Cmd) WithDir(dir string) CmdIface {
	c.cmd.Dir = dir
	return c
}

// WithStdout sets the stdout for the command.
func (c *Cmd) WithStdout(stdout *os.File) CmdIface {
	c.cmd.Stdout = stdout
	return c
}

// WithStderr sets the stderr for the command.
func (c *Cmd) WithStderr(stderr *os.File) CmdIface {
	c.cmd.Stderr = stderr
	return c
}

// Run executes the command.
func (c *Cmd) Run() error {
	return c.cmd.Run()
}

// Output runs the command and returns its standard output.
func (c *Cmd) Output() ([]byte, error) {
	return c.cmd.Output()
}

// CombinedOutput runs the command and returns its combined standard output and standard error.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	return c.cmd.CombinedOutput()
}

// NewExec creates a real executor.
func NewExec() ExecIface {
	return Exec{}
}
