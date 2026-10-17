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
- **internal/gh/** - GitHub API client wrapper using `cli/go-gh/v2`; split across five files:
  - `gh.go` — `CurrentLogin()`, generic `Search[T]` (runs parallel GraphQL queries), `sortedKeys` (deterministic iteration over the per-query result map)
  - `pr.go` — `SearchPRs`, `SearchPRsTeams`, `MergeSearchPRsResults`, `PRSearchNode`, `PRSearchResult`
  - `issue.go` — `SearchIssues`, `SearchIssuesTeams`, `MergeSearchIssuesResults`, `IssueSearchNode`, `IssueSearchResult`
  - `activity.go` — `LatestActivity`, `NewLatestActivity` (picks most recent of comment / review / push)
  - `conversation.go` — `Conversation` (recent comments, reviews, review threads, commits of a PR, fetched by the PR search query), `lastActivityBy`, and `horizon` (truncation horizon: events older than the oldest entry of a full list are not judged)
  - `attention.go` — `Conversation.Attention(login, owned)` decides whether a PR is waiting on the user (mentioned / replied to you / commented) and `mentions` (whole-login @-match; skips code spans, fenced blocks, and quoted lines)
  - `sort.go` — `SortAt` (the timestamp a list entry is ordered by), `SortByUpdatedDesc`
- **internal/pr/** - PR data types and search logic (groups by: needs action, ready to merge, review-requested, waiting, drafts, participated); `NewGroupedPullRequests(result, login, conversation bool)` runs `needsaction.go`'s promotion when `conversation` is true: PRs whose `Conversation.Attention` fires move into Needs action and leave the other default tabs (Review Requested keeps them); `BuildTabs()` produces `[]ui.Tab`
- **internal/issue/** - Issue data types and search logic (groups by: created, assigned, participated); `BuildTabs()` produces `[]ui.Tab`
- **internal/ui/** - Bubbletea TUI with tabbed interface; `Model` manages tabs, `Item` represents list entries, `NewLoadingModel` shows a spinner while data is fetched
- **internal/config/** - YAML config at `~/.config/gh-own/config.yaml`: per-tab queries, `exclude`, and `pr.needsAction.conversation` (default true; false skips both the conversation fields in the PR search query and `promoteNeedsAction`, so every tab is exactly its query)
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
4. Domain packages (`pr`, `issue`) group results, order every bucket newest first by `gh.SortAt`, and call `BuildTabs()` to produce `[]ui.Tab`; `pr.NewGroupedPullRequests` first runs `promoteNeedsAction` (unless `pr.needsAction.conversation` is false), so Needs action = changes requested ∪ PRs waiting on the user (see `gh.Conversation.Attention`); the description line and sort time of such a PR come from `PRSearchNode.Attention` instead of `LatestActivity`
5. `ui.NewLoadingModel` starts Bubbletea with a spinner; `ui.FetchCmd` wraps the fetch function and delivers `TabsMsg` (success) or `ErrMsg` (failure)
6. Keyboard: `enter` opens the selected item URL in the system browser; `r` refreshes; `s` toggles newest/oldest first (`Model.applySort` re-sorts every tab and is re-applied on `TabsMsg`); `tab`/`shift+tab` switch tabs; `/` filters

Search queries carry `sort:updated-desc` unless the user's own query specifies a `sort:` — `config.PRSearchEntries` / `config.IssueSearchEntries` build the user queries, `gh.prTeamEntries` / `gh.issueTeamEntries` the team ones.
