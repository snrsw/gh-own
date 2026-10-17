---
type: Pipeline
title: Search pipeline
description: How a command turns config and flags into GitHub search queries, runs user and team searches concurrently, and merges the results.
resource: https://github.com/snrsw/gh-own/blob/main/cmd/pr.go
tags: [search, graphql, teams, filters, cache]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Stages

`cmd/pr.go` and `cmd/issue.go` follow the same steps:

1. Load the [config file](/config/config-file.md).
2. `gh.CurrentLogin()` reads the login from the `gh` CLI config for the default host.
3. Build the user queries with `config.PRSearchEntries` / `config.IssueSearchEntries` (see [query building](#query-building)).
4. Run two goroutines at the same time:
   - **User search** — `gh.SearchPRs` / `gh.SearchIssues` over the user queries.
   - **Team search** — `gh.GetTeamSlugsWithCache` (REST `user/teams?per_page=50`), then one `team:<org>/<slug>` query per team, keyed `participatedTeam<N>`.
5. Merge with `gh.MergeSearchPRsResults` / `gh.MergeSearchIssuesResults`; every bucket is deduplicated by URL.
6. Group with `pr.NewGroupedPullRequests` / `issue.NewGroupedIssues`, then `BuildTabs()`.

An error from either goroutine aborts the whole fetch. There is no partial result.

`gh.Search` runs **one GraphQL request per query, all in parallel**, each with `search(type: ISSUE, first: 50)`. Each tab therefore sees at most 50 results per query.

# Query building

`searchEntries` applies these steps to every user query:

| Step | Function | Effect |
|---|---|---|
| Merge | `MergePRQueries` / `MergeIssueQueries` | Config queries override defaults key by key. Aliases `participated` → `participatedUser` and `review_requested` → `reviewRequested` are normalized at load time. |
| `{user}` | `ResolveQueries` | Replaced with the login. |
| `{owner}` | `ResolveQueries` | Splits one query into two: `key` with `author:<login>` and `keyAssigned` with `assignee:<login>`. GitHub search cannot express "author OR assignee". Both keys route to the same tab by prefix (see [PR tabs](/tabs/pr-tabs.md)). |
| Exclusions | `ExcludeAuthors` | Appends `-author:<login>` per excluded login, unless the query already names that author in an `author:` or `-author:` field. |
| Org | `AppendOrg` | Appends `org:<org>` when `--org` is set. |
| Sort | `EnsureSort` | Appends `sort:updated-desc` unless the query has a `sort:` field. |

Team queries skip the merge and placeholder steps but go through the same `config.ApplyFilters` (exclusions, org, sort). **Any new filter must be added to `ApplyFilters`** so it reaches team tabs too.

`sort:updated-desc` matters because of the 50-result cap. Without it, GitHub returns the 50 most *relevant* results, not the 50 most recent.

## Author exclusions

`ResolveExcludeAuthors` unions three sources in this order: the bot preset (`exclude.bots` or `--no-bots` → `app/dependabot`, `app/renovate`, `app/github-actions`), `exclude.authors`, and `--exclude-author`. Duplicates collapse by `canonicalAuthor`, which maps `app/renovate` and `renovate[bot]` to `renovate`.

Exclusion happens in the query rather than client-side, so excluded items do not use up any of the 50 result slots.

# Team cache

Team slugs are cached at `<gh cache dir>/gh-own/teams.json` with a 6-hour TTL (`internal/cache`). The rules:

- A missing, unreadable, corrupt, or expired file is a cache miss, never an error.
- Writes go to `teams.json.tmp`, then rename, so a crash never leaves a half-written file.
- A failed write is ignored; the fetched teams are still used.

# Demo mode

`--demo` skips every stage above and feeds `internal/demo` fixtures into the grouping with an empty login. An empty login disables [Needs action promotion](/rules/needs-action-promotion.md), so the fixtures render exactly as written.

# Debugging

`--debug` enables `slog` debug output. `timing.Track("<cmd>:<stage>")` logs the duration of each stage above.
