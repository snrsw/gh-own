// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
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

		client, err := api.DefaultGraphQLClient()
		if err != nil {
			return err
		}

		store, err := cache.NewStore()
		if err != nil {
			return err
		}

		userCh := make(chan result[*gh.PRSearchResult], 1)
		go func() {
			prs, err := gh.SearchPRs(client, username)
			userCh <- result[*gh.PRSearchResult]{v: prs, err: err}
		}()

		teamCh := make(chan result[*gh.PRSearchResult], 1)
		go func() {
			teams, err := gh.GetTeamSlugsWithCache(restClient, store, 6*time.Hour)
			if err != nil {
				teamCh <- result[*gh.PRSearchResult]{v: nil, err: err}
				return
			}

			prs, err := gh.SearchPRsTeams(client, username, teams)
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

		prs := gh.MergeSearchPRsResults(userResult.v, teamResult.v)
		prg := pr.NewGroupedPullRequests(prs)

		return prg.View()
	},
}
