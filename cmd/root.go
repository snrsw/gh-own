// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/snrsw/gh-own/internal/config"
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
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		if debug {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
		return nil
	},
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

var debug bool
var demo bool
var org string
var excludeAuthors []string
var noBots bool

func init() {
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&demo, "demo", false, "use demo data (no GitHub API calls)")
	rootCmd.PersistentFlags().StringVar(&org, "org", "", "filter by organization")
	rootCmd.PersistentFlags().StringSliceVar(&excludeAuthors, "exclude-author", nil, "hide items written by this author (repeatable)")
	rootCmd.PersistentFlags().BoolVar(&noBots, "no-bots", false, "hide items written by common bots (dependabot, renovate, github-actions)")
	rootCmd.AddCommand(prCmd, issueCmd)
}

// searchFilters folds the exclusion sources into the filters every search
// query is narrowed by.
func searchFilters(cfg config.Config) config.Filters {
	return config.Filters{
		Org:            org,
		ExcludeAuthors: config.ResolveExcludeAuthors(cfg.Exclude, excludeAuthors, noBots),
	}
}
