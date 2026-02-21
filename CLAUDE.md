# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```sh
go build              # Build the extension
go test ./...         # Run all tests
go test -v ./...      # Run tests with verbose output
go test ./internal/ui # Run tests for a specific package
golangci-lint run     # Run linter
```

## Architecture

gh-own is a GitHub CLI extension that displays the user's owned PRs and issues in a terminal UI.

### Package Structure

- **cmd/** - Cobra command definitions (`root`, `pr`, `issue`); `root` has a persistent `--debug` flag that enables slog debug output
- **internal/gh/** - GitHub API client wrapper using `cli/go-gh/v2`; split across four files:
  - `gh.go` — `CurrentLogin()`, generic `Search[T]` (runs parallel GraphQL queries)
  - `pr.go` — `SearchPRs`, `SearchPRsTeams`, `MergeSearchPRsResults`, `PRSearchNode`, `PRSearchResult`
  - `issue.go` — `SearchIssues`, `SearchIssuesTeams`, `MergeSearchIssuesResults`, `IssueSearchNode`, `IssueSearchResult`
  - `activity.go` — `LatestActivity`, `NewLatestActivity` (picks most recent of comment / review / push)
- **internal/pr/** - PR data types and search logic (groups by: created, assigned, review-requested, participated); `BuildTabs()` produces `[]ui.Tab`
- **internal/issue/** - Issue data types and search logic (groups by: created, assigned, participated); `BuildTabs()` produces `[]ui.Tab`
- **internal/ui/** - Bubbletea TUI with tabbed interface; `Model` manages tabs, `Item` represents list entries, `NewLoadingModel` shows a spinner while data is fetched
- **internal/cache/** - Team slug cache stored at `~/.cache/gh/gh-own/teams.json`; written atomically; TTL-based expiry (default 6 h)
- **internal/cistatus/** - `CIStatus` enum (None/Success/Failure/Pending) parsed from GitHub's `statusCheckRollup`; `RenderCIStatus` returns coloured symbol
- **internal/reviewstatus/** - `ReviewStatus` enum (None/Approved/ChangesRequested/ReviewRequired) parsed from GitHub's `reviewDecision`; `RenderReviewStatus` returns coloured symbol
- **internal/timing/** - `Track(name string) func()` deferred helper that logs stage duration via `slog.Debug`

### Data Flow

1. Commands call `gh.CurrentLogin()` to get the authenticated username
2. Two goroutines run concurrently:
   - **User search** — `gh.SearchPRs` / `gh.SearchIssues` via GraphQL
   - **Team search** — `gh.GetTeamSlugsWithCache` (REST, cached 6 h) → `gh.SearchPRsTeams` / `gh.SearchIssuesTeams`
3. Results are merged with `gh.MergeSearch*Results` (deduplicates by URL)
4. Domain packages (`pr`, `issue`) group results and call `BuildTabs()` to produce `[]ui.Tab`
5. `ui.NewLoadingModel` starts Bubbletea with a spinner; `ui.FetchCmd` wraps the fetch function and delivers `TabsMsg` (success) or `ErrMsg` (failure)
6. Keyboard: `enter` opens the selected item URL in the system browser; `r` refreshes; `tab`/`shift+tab` switch tabs; `/` filters
