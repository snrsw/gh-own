// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/pullrequest"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "GitHub CLI extension to list your owned pull requests.",
	Long:  "GitHub CLI extension to list your owned pull requests.",
	RunE: func(_ *cobra.Command, _ []string) error {
		client, username, err := gh.CurrentUser()
		if err != nil {
			return err
		}

		prs, err := pullrequest.SearchPullRequests(client, username)
		if err != nil {
			return err
		}

		return prs.View()
	},
}
