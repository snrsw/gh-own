// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type result[T any] struct {
	v   T
	err error
}

var rootCmd = &cobra.Command{
	Use:   "gh-own",
	Short: "GitHub CLI extension to list your owned pull requests and issues.",
	Long:  "GitHub CLI extension to list your owned pull requests and issues. If no subcommand is specified, the pr subcommand is run by default.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return prCmd.RunE(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(prCmd, issueCmd)
}
