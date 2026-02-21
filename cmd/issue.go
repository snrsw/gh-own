// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/issue"
	"github.com/snrsw/gh-own/internal/timing"
	"github.com/snrsw/gh-own/internal/ui"
	"github.com/spf13/cobra"
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "GitHub CLI extension to list your owned issues.",
	Long:  "GitHub CLI extension to list your owned issues.",
	RunE: func(_ *cobra.Command, _ []string) error {
		defer timing.Track("issue:total")()

		fetch := ui.FetchCmd(func() ([]ui.Tab, error) {
			done := timing.Track("issue:login")
			username, err := gh.CurrentLogin()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("issue:rest-client")
			restClient, err := api.DefaultRESTClient()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("issue:graphql-client")
			client, err := api.DefaultGraphQLClient()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("issue:cache-store")
			store, err := cache.NewStore()
			done()
			if err != nil {
				return nil, err
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
				return nil, userResult.err
			}

			teamResult := <-teamCh
			if teamResult.err != nil {
				return nil, teamResult.err
			}

			done = timing.Track("issue:merge-results")
			issues := gh.MergeSearchIssuesResults(userResult.v, teamResult.v)
			done()

			done = timing.Track("issue:group")
			ig := issue.NewGroupedIssues(issues)
			done()

			return ig.BuildTabs(), nil
		})

		m := ui.NewLoadingModel(fetch)
		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}
		if fm, ok := finalModel.(ui.Model); ok {
			if fmErr := fm.Err(); fmErr != nil {
				return fmErr
			}
		}
		return nil
	},
}
