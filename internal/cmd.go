package internal

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"os/exec"

	"al.essio.dev/pkg/shellescape"
)

type Cmd struct {
	shell   string
	command string
	Dir     string
}

func NewCmd(shell, command string) *Cmd {
	return &Cmd{
		shell:   shell,
		command: command,
	}
}

func (c Cmd) Arg(a string) *Cmd {
	return &Cmd{
		shell:   c.shell,
		command: c.command + " " + shellescape.Quote(a),
	}
}

type CmdResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Err      error
}

// Run command by `shell -c command` style.
func (c Cmd) Run(ctx context.Context) *CmdResult {
	cmd := exec.CommandContext(ctx, c.shell, "-c", c.command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = c.Dir

	slog.Debug("Start command", slog.Any("args", cmd.Args))

	var r CmdResult
	r.Err = cmd.Run()
	if exitErr, ok := errors.AsType[*exec.ExitError](r.Err); ok {
		r.ExitCode = exitErr.ExitCode()
	}
	r.Stdout = stdout.String()
	r.Stderr = stderr.String()

	slog.Debug("End command", slog.Any("args", cmd.Args), slog.Int("exit", r.ExitCode))
	return &r
}
