// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/issue"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "GitHub CLI extension to list your owned issues.",
	Long:  "GitHub CLI extension to list your owned issues.",
	RunE: func(_ *cobra.Command, _ []string) error {
		username, err := gh.CurrentLogin()
		if err != nil {
			return err
		}

		client, err := api.DefaultGraphQLClient()
		if err != nil {
			return err
		}

		issues, err := issue.SearchIssues(client, username)
		if err != nil {
			return err
		}

		return issues.View()
	},
}
