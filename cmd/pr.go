// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/pr"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "GitHub CLI extension to list your owned pull requests.",
	Long:  "GitHub CLI extension to list your owned pull requests.",
	RunE: func(_ *cobra.Command, _ []string) error {
		username, err := gh.CurrentLogin()
		if err != nil {
			return err
		}

		restClient, err := api.DefaultRESTClient()
		if err != nil {
			return err
		}

		teams, err := gh.GetTeamSlugs(restClient)
		if err != nil {
			return err
		}

		client, err := api.DefaultGraphQLClient()
		if err != nil {
			return err
		}

		prs, err := pr.SearchPullRequests(client, username, teams)
		if err != nil {
			return err
		}

		return prs.View()
	},
}
