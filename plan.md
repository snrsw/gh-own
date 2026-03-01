# Custom Tab Configuration

## Context

gh-own currently has fixed tabs (PR: Created, Participated, Assigned, Review Requested; Issue: Created, Participated, Assigned). Users want to add custom tabs with arbitrary GitHub search queries via `config.yaml`. The config and search layers already support arbitrary keys — the bottleneck is the parse→group→tab pipeline that drops unknown keys.

## Design

Any non-default key in `queries` becomes a custom tab. Tab name = config key. Custom tabs appear after default tabs, sorted alphabetically by key.

```yaml
pr:
  queries:
    needs-triage: "is:pr is:open label:needs-triage"
issue:
  queries:
    bugs: "is:issue is:open label:bug"
```

## Files to Modify

- `internal/config/config.go` — add `DefaultPRKeys()`, `DefaultIssueKeys()`
- `internal/gh/pr.go` — add `Custom` field to `PRSearchResult`, update `parsePRSearchResult`, `MergeSearchPRsResults`
- `internal/gh/issue.go` — same for `IssueSearchResult`
- `internal/pr/pr.go` — add `Custom` to `GroupedPullRequests`, update `NewGroupedPullRequests`
- `internal/pr/ui.go` — update `BuildTabs()` to append sorted custom tabs
- `internal/issue/issue.go` — add `Custom` to `GroupedIssues`, update `NewGroupedIssues`
- `internal/issue/ui.go` — update `BuildTabs()` to append sorted custom tabs

No changes needed in `cmd/` — data flows through naturally.

## Test Plan (TDD)

### Config layer (`internal/config/config_test.go`)

- [x] `TestDefaultPRKeys_ReturnsKnownKeys` — returns set of {created, assigned, participatedUser, reviewRequested}
- [ ] `TestDefaultIssueKeys_ReturnsKnownKeys` — returns set of {created, assigned, participatedUser}

### Search layer — PR (`internal/gh/pr_test.go`)

- [ ] `TestParsePRSearchResult_CustomKeyPreserved` — unknown key `"myTab"` lands in `result.Custom["myTab"]`
- [ ] `TestParsePRSearchResult_NoCustomKeys` — standard keys only → `Custom` is empty map
- [ ] `TestMergeSearchPRsResults_MergesCustom` — merges custom maps, concatenates overlapping keys

### Search layer — Issue (`internal/gh/issue_test.go`)

- [ ] `TestParseIssueSearchResult_CustomKeyPreserved` — same for issues
- [ ] `TestParseIssueSearchResult_NoCustomKeys` — same for issues
- [ ] `TestMergeSearchIssuesResults_MergesCustom` — same for issues

### Domain layer — PR (`internal/pr/pr_test.go`)

- [ ] `TestNewGroupedPullRequests_PropagatesCustom` — custom nodes convert to `SearchResult[pullRequest]`
- [ ] `TestBuildTabs_DefaultTabsOnly` — no custom → exactly 4 tabs
- [ ] `TestBuildTabs_WithCustomTabs` — custom keys "zeta" and "alpha" → 6 tabs, custom sorted alphabetically at indices 4-5
- [ ] `TestBuildTabs_CustomTabNameIncludesCount` — tab name format: `"myTab (3)"`

### Domain layer — Issue (`internal/issue/issue_test.go`)

- [ ] `TestNewGroupedIssues_PropagatesCustom` — same for issues
- [ ] `TestBuildTabs_Issue_DefaultTabsOnly` — no custom → exactly 3 tabs
- [ ] `TestBuildTabs_Issue_WithCustomTabs` — custom tabs appended after 3 defaults, sorted

### Existing test update (`internal/gh/pr_test.go`)

- [ ] `TestSearchPRs_EmptyEntries_HasEmptyCustom` — assert `Custom` is not nil, length 0

## Verification

```sh
go test ./...
golangci-lint run
```
