---
type: TabSet
title: PR tabs
description: The six default tabs of `gh-own pr`, the query behind each, and how query keys route results into tab buckets.
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/pr.go
tags: [pr, tabs, query-keys]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Tabs

Tabs render in this order (`internal/pr/ui.go`, `BuildTabs`):

| Tab | Bucket field | Query key | Default query | Owned? |
|---|---|---|---|---|
| Needs action | `NeedsAction` | `needsAction` | `is:pr is:open draft:false review:changes-requested {owner}` | yes |
| Ready to merge | `ReadyToMerge` | `readyToMerge` | `is:pr is:open draft:false review:approved {owner}` | yes |
| Review Requested | `ReviewRequested` | `reviewRequested` | `is:pr is:open review-requested:{user}` | no |
| Waiting for review or checks | `Waiting` | `waiting` | `is:pr is:open draft:false -review:approved -review:changes-requested {owner}` | yes |
| Drafts | `Drafts` | `drafts` | `is:pr is:open draft:true {owner}` | yes |
| Participated | `Participated` | `participatedUser` + `participatedTeam<N>` | `is:pr is:open involves:{user} -author:{user} -assignee:{user} -review-requested:{user}` | no |
| Custom (one per extra key) | `Custom[key]` | any non-default key | from config | — |

"Owned" means the query uses `{owner}`, so the PR is authored by or assigned to the user. The [attention rule](/rules/attention.md) treats owned PRs differently.

Custom tabs follow the defaults, sorted by key; the tab name comes from `ui.HumanizeTabName(key)`.

# Key routing

`parsePRSearchResult` assigns each query result to a bucket by **key prefix**, not by exact key:

- `drafts*`, `needsAction*`, `readyToMerge*`, `waiting*`, `participated*` → the matching bucket.
- `reviewRequested` → exact match only.
- Any other key that is not a default key → `Custom[key]`.

Prefix routing is load-bearing. Two kinds of generated keys depend on it:

- `{owner}` expansion adds a `<key>Assigned` sibling query (see [search pipeline](/pipelines/search.md)).
- Team searches use keys `participatedTeam0`, `participatedTeam1`, …

Each bucket is deduplicated by PR URL, and keys are iterated through `sortedKeys` so the result is deterministic.

**When adding a default PR tab, edit four places together:** `defaultPRQueries` in `internal/config/config.go`, the bucket switch in `parsePRSearchResult`, `PRSearchResult` / `MergeSearchPRsResults`, and `BuildTabs`. A key accepted by none of the prefixes silently becomes a custom tab.

# After grouping

`pr.NewGroupedPullRequests` may move PRs between buckets before rendering; see [Needs action promotion](/rules/needs-action-promotion.md). Every bucket is then sorted newest first; see [activity line and sort order](/ui/activity-and-sort.md).
