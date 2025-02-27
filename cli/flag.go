package cli

import (
	"strconv"
	"time"

	"github.com/ryota2357/looprun/runner"
	"github.com/spf13/cobra"
)

type Flag struct {
	quiet       bool
	noStdout    bool
	noStderr    bool
	stopOnError bool
	stopOnCode  int
	timeout     string
	maxRuns     int
}

func initFlags(cmd *cobra.Command, flags *Flag) {
	cmd.Flags().SetInterspersed(false)

	cmd.Flags().BoolVarP(
		&flags.quiet, "quiet", "q", false,
		"Suppress output from stdout and stderr. This option is equivalent to --no-stdout --no-stderr.",
	)
	cmd.Flags().BoolVarP(
		&flags.noStdout, "no-stdout", "O", false,
		"Suppress output from stdout.",
	)
	cmd.Flags().BoolVarP(&flags.noStderr, "no-stderr", "E", true,
		"Suppress output from stderr.",
	)
	cmd.Flags().BoolVarP(&flags.stopOnError, "stop-on-error", "e", false,
		"Stop execution when the command returns a non-zero exit code.",
	)
	cmd.Flags().IntVarP(&flags.stopOnCode, "stop-on-code", "c", -1,
		"Stop execution when the command exit with the specified exit code.",
	)
	cmd.Flags().StringVarP(&flags.timeout, "timeout", "t", "-1",
		"Stop execution when the command takes longer than the specified time.",
	)
	cmd.Flags().IntVarP(&flags.maxRuns, "max-runs", "m", -1,
		"Stop execution after the specified number of runs.",
	)

}

func createConfig(flags *Flag) (runner.Config, error) {

	var isStopExitCode func(code int) bool
	if flags.stopOnError {
		isStopExitCode = func(code int) bool { return code != 0 }
	} else if flags.stopOnCode >= 0 {
		stopCode := flags.stopOnCode
		isStopExitCode = func(code int) bool { return code == stopCode }
	} else {
		isStopExitCode = func(code int) bool { return false }
	}

	var timeout time.Duration
	if seconds, err := strconv.ParseFloat(flags.timeout, 64); err == nil {
		timeout = time.Duration(seconds * float64(time.Second))
	} else {
		timeout, err = time.ParseDuration(flags.timeout)
		if err != nil {
			return runner.Config{}, err
		}
	}

	config := runner.Config{
		IsStopExitCode: isStopExitCode,
		Timeout:        timeout,
		StdoutEnabled:  !flags.quiet && !flags.noStdout,
		StderrEnabled:  !flags.quiet && !flags.noStderr,
		MaxRuns:        flags.maxRuns,
	}
	return config, nil
}
