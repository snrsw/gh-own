// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/pr"
	"github.com/snrsw/gh-own/internal/timing"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "GitHub CLI extension to list your owned pull requests.",
	Long:  "GitHub CLI extension to list your owned pull requests.",
	RunE: func(_ *cobra.Command, _ []string) error {
		defer timing.Track("pr:total")()

		done := timing.Track("pr:login")
		username, err := gh.CurrentLogin()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("pr:rest-client")
		restClient, err := api.DefaultRESTClient()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("pr:graphql-client")
		client, err := api.DefaultGraphQLClient()
		done()
		if err != nil {
			return err
		}

		done = timing.Track("pr:cache-store")
		store, err := cache.NewStore()
		done()
		if err != nil {
			return err
		}

		userCh := make(chan result[*gh.PRSearchResult], 1)
		go func() {
			defer timing.Track("pr:search-user")()
			prs, err := gh.SearchPRs(client, username)
			userCh <- result[*gh.PRSearchResult]{v: prs, err: err}
		}()

		teamCh := make(chan result[*gh.PRSearchResult], 1)
		go func() {
			defer timing.Track("pr:search-teams-total")()

			teamDone := timing.Track("pr:get-team-slugs")
			teams, err := gh.GetTeamSlugsWithCache(restClient, store, 6*time.Hour)
			teamDone()
			if err != nil {
				teamCh <- result[*gh.PRSearchResult]{v: nil, err: err}
				return
			}

			teamDone = timing.Track("pr:search-teams")
			prs, err := gh.SearchPRsTeams(client, username, teams)
			teamDone()
			if err != nil {
				teamCh <- result[*gh.PRSearchResult]{v: nil, err: err}
				return
			}
			teamCh <- result[*gh.PRSearchResult]{v: prs, err: err}
		}()

		userResult := <-userCh
		if userResult.err != nil {
			return userResult.err
		}

		teamResult := <-teamCh
		if teamResult.err != nil {
			return teamResult.err
		}

		done = timing.Track("pr:merge-results")
		prs := gh.MergeSearchPRsResults(userResult.v, teamResult.v)
		done()

		done = timing.Track("pr:group")
		prg := pr.NewGroupedPullRequests(prs)
		done()

		return prg.View()
	},
}
