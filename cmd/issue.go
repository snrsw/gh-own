// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
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

		restClient, err := api.DefaultRESTClient()
		if err != nil {
			return err
		}

		client, err := api.DefaultGraphQLClient()
		if err != nil {
			return err
		}

		store, err := cache.NewStore()
		if err != nil {
			return err
		}

		userCh := make(chan result[*gh.IssueSearchResult], 1)
		go func() {
			issues, err := gh.SearchIssues(client, username)
			userCh <- result[*gh.IssueSearchResult]{v: issues, err: err}
		}()

		teamCh := make(chan result[*gh.IssueSearchResult], 1)
		go func() {
			teams, err := gh.GetTeamSlugsWithCache(restClient, store, 6*time.Hour)
			if err != nil {
				teamCh <- result[*gh.IssueSearchResult]{v: nil, err: err}
				return
			}

			issues, err := gh.SearchIssuesTeams(client, username, teams)
			if err != nil {
				teamCh <- result[*gh.IssueSearchResult]{v: nil, err: err}
				return
			}
			teamCh <- result[*gh.IssueSearchResult]{v: issues, err: err}
		}()

		userResult := <-userCh
		if userResult.err != nil {
			return userResult.err
		}

		teamResult := <-teamCh
		if teamResult.err != nil {
			return teamResult.err
		}

		issues := gh.MergeSearchIssuesResults(userResult.v, teamResult.v)
		ig := issue.NewGroupedIssues(issues)

		return ig.View()
	},
}
