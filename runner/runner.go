package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var CommandNotFound = errors.New("command not found")

var FailedToRunCommand = faildToRunCommand{}

type faildToRunCommand struct{ err error }

func (e *faildToRunCommand) Error() string { return fmt.Sprintf("failed to run command: %v", e.err) }
func (e *faildToRunCommand) Unwrap() error { return e.err }
func (e *faildToRunCommand) Is(target error) bool {
	_, ok := target.(*faildToRunCommand)
	return ok
}

type TerminationReason int

const (
	TerminatedByExitCode TerminationReason = iota
	TerminatedByMaxRuns
	TerminatedByTimeout
	TerminatedByUnknown
)

func (r TerminationReason) String() string {
	switch r {
	case TerminatedByExitCode:
		return "exit code"
	case TerminatedByMaxRuns:
		return "max runs"
	case TerminatedByTimeout:
		return "timeout"
	default:
		return "unknown"
	}
}

type Output struct {
	Stdout      []byte
	Stderr      []byte
	ExitCode    int
	Termination TerminationReason
}

type Config struct {
	IsStopExitCode func(code int) bool
	StdoutEnabled  bool
	StderrEnabled  bool
	Timeout        time.Duration
	MaxRuns        int
	Interval       time.Duration
}

func New(name string, args []string, config Config) *Runner {
	return &Runner{
		commandName: name,
		commandArgs: args,
		config:      config,
	}
}

type Runner struct {
	commandName string
	commandArgs []string
	config      Config
}

func (r *Runner) Run() (Output, error) {
	if _, err := exec.LookPath(r.commandName); err != nil {
		return Output{}, CommandNotFound
	}

	var cmdCtx context.Context
	var cancel context.CancelFunc
	if r.config.Timeout > 0 {
		cmdCtx, cancel = context.WithTimeout(context.Background(), r.config.Timeout)
	} else {
		cmdCtx, cancel = context.WithCancel(context.Background())
	}
	defer cancel()

	output := Output{}
	runCount := 0
	var lastStdout, lastStderr []byte

	for {
		if r.config.MaxRuns > 0 && runCount >= r.config.MaxRuns {
			output.Termination = TerminatedByMaxRuns
			break
		}

		runCount += 1
		startTime := time.Now()

		exitCode, stdout, stderr, err := r.runIteration(cmdCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				output.Termination = TerminatedByTimeout
				output.ExitCode = -1
				lastStdout = stdout
				lastStderr = stderr
				break
			}
			return Output{}, &faildToRunCommand{err}
		}

		output.ExitCode = exitCode
		lastStdout = stdout
		lastStderr = stderr

		if r.config.Interval > 0 {
			elapsed := time.Since(startTime)
			remaing := r.config.Interval - elapsed
			if remaing > 0 {
				select {
				case <-time.After(remaing):
				case <-cmdCtx.Done():
					output.Termination = TerminatedByTimeout
					break
				}
			}
		}

		if cmdCtx.Err() != nil {
			output.Termination = TerminatedByTimeout
			output.ExitCode = -1
			break
		}

		if r.config.IsStopExitCode(exitCode) {
			output.Termination = TerminatedByExitCode
			break
		}

	}

	if !r.config.StdoutEnabled {
		output.Stdout = lastStdout
	}
	if !r.config.StderrEnabled {
		output.Stderr = lastStderr
	}

	return output, nil
}

func (r *Runner) runIteration(ctx context.Context) (int, []byte, []byte, error) {
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, r.commandName, r.commandArgs...)

	if r.config.StdoutEnabled {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &stdoutBuf
	}
	if r.config.StderrEnabled {
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stderr = &stderrBuf
	}

	err := cmd.Run()

	stdout := stdoutBuf.Bytes()
	stderr := stderrBuf.Bytes()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return -1, stdout, stderr, err
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), stdout, stderr, nil
		}
		return -1, stdout, stderr, err
	}

	return 0, stdout, stderr, nil
}
