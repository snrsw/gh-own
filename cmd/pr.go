// Package cmd implements gh-own CLI subcommands.
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/snrsw/gh-own/internal/cache"
	"github.com/snrsw/gh-own/internal/checkout"
	"github.com/snrsw/gh-own/internal/gh"
	"github.com/snrsw/gh-own/internal/pr"
	"github.com/snrsw/gh-own/internal/timing"
	"github.com/snrsw/gh-own/internal/ui"
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "GitHub CLI extension to list your owned pull requests.",
	Long:  "GitHub CLI extension to list your owned pull requests.",
	RunE: func(_ *cobra.Command, _ []string) error {
		defer timing.Track("pr:total")()

		fetch := ui.FetchCmd(func() ([]ui.Tab, error) {
			done := timing.Track("pr:login")
			username, err := gh.CurrentLogin()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("pr:rest-client")
			restClient, err := api.DefaultRESTClient()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("pr:graphql-client")
			client, err := api.DefaultGraphQLClient()
			done()
			if err != nil {
				return nil, err
			}

			done = timing.Track("pr:cache-store")
			store, err := cache.NewStore()
			done()
			if err != nil {
				return nil, err
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
				return nil, userResult.err
			}

			teamResult := <-teamCh
			if teamResult.err != nil {
				return nil, teamResult.err
			}

			done = timing.Track("pr:merge-results")
			prs := gh.MergeSearchPRsResults(userResult.v, teamResult.v)
			done()

			done = timing.Track("pr:group")
			prg := pr.NewGroupedPullRequests(prs)
			done()

			return prg.BuildTabs(), nil
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
			if repo, number, ok := fm.CheckoutRequest(); ok {
				return runCheckout(repo, number)
			}
		}
		return nil
	},
}

func runCheckout(repo string, number int) error {
	repoDir, err := checkout.FindRepoDir(repo, ghqOutput)
	if err != nil {
		return fmt.Errorf("finding repo: %w", err)
	}
	return checkout.Checkout(repoDir, number, execRun)
}

func ghqOutput(name string, args []string, _ string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}

func execRun(name string, args []string, dir string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
