// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/issue"
	"github.com/snrsw/gh-own/internal/timing"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "GitHub CLI extension to list your owned issues.",
	Long:  "GitHub CLI extension to list your owned issues.",
	RunE: func(_ *cobra.Command, _ []string) error {
		defer timing.Track("issue:total")()

		done := timing.Track("issue:login")
		username, err := gh.CurrentLogin()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("issue:rest-client")
		restClient, err := api.DefaultRESTClient()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("issue:graphql-client")
		client, err := api.DefaultGraphQLClient()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("issue:cache-store")
		store, err := cache.NewStore()
		done()
		if err != nil {
			return err
		}

		userCh := make(chan result[*gh.IssueSearchResult], 1)
		go func() {
			defer timing.Track("issue:search-user")()
			issues, err := gh.SearchIssues(client, username)
			userCh <- result[*gh.IssueSearchResult]{v: issues, err: err}
		}()

		teamCh := make(chan result[*gh.IssueSearchResult], 1)
		go func() {
			defer timing.Track("issue:search-teams-total")()

			teamDone := timing.Track("issue:get-team-slugs")
			teams, err := gh.GetTeamSlugsWithCache(restClient, store, 6*time.Hour)
			teamDone()
			if err != nil {
				teamCh <- result[*gh.IssueSearchResult]{v: nil, err: err}
				return
			}

			teamDone = timing.Track("issue:search-teams")
			issues, err := gh.SearchIssuesTeams(client, username, teams)
			teamDone()
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

		done = timing.Track("issue:merge-results")
		issues := gh.MergeSearchIssuesResults(userResult.v, teamResult.v)
		done()

		done = timing.Track("issue:group")
		ig := issue.NewGroupedIssues(issues)
		done()

		return ig.View()
	},
}
