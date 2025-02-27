package cli

import (
	"fmt"

	"github.com/ryota2357/looprun/runner"
	"github.com/spf13/cobra"
)

func Execute() error {
	var flag Flag
	var config runner.Config
	rootCmd := &cobra.Command{
		Use:     "looprun [flags] [command [command args...]]",
		Short:   "Repeat a given command.",
		Long:    ``,
		Version: "0.0.0",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			config, err = createConfig(&flag)
			if err != nil {
				return err
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			r := runner.New(
				args[0],
				args[1:],
				config,
			)
			output, err := r.Run()
			if err != nil {
				fmt.Println(err)
			} else {
				// fmt.Println(string(output.Stdout))
				// fmt.Println(string(output.Stderr))
				fmt.Println(output.ExitCode)
				fmt.Println(output.Termination)
			}
		},
	}
	initFlags(rootCmd, &flag)

	return rootCmd.Execute()
}
