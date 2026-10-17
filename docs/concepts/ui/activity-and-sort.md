---
type: DisplayRule
title: Activity line and sort order
description: What the second line of each list item reports, and why the list is ordered by exactly that timestamp.
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/sort.go
tags: [ui, sort, activity]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Invariant

**The time a list is sorted by is the time shown on the description line.** `gh.SortAt(activity, updatedAt)` returns the activity time when the activity has a login, and `updatedAt` otherwise. Each item stores this value as `SortAt`.

# Which activity is shown

For PRs, `pullRequest.shownActivity()` decides:

1. If `Attention.Reason` is set (see [attention](/rules/attention.md)), show it: `mentioned you by @alice 2h ago · opened on … by …`.
2. Otherwise show `LatestActivity`.
3. If that has no login, show `updated 2h ago · opened on … by …`.

`LatestActivity` (`internal/gh/activity.go`) is the most recent of three candidates:

| Candidate | Source | Kind text |
|---|---|---|
| Last comment | `Conversation.Comments` | `commented` |
| Last review | `Conversation.Reviews` | `approved`, `changes requested`, `commented`, or `dismissed` |
| Last push | `Conversation.Commits` (only when the login is known) | `pushed` |

For issues, only the last comment is a candidate.

# Ordering

- `gh.SortByUpdatedDesc` sorts each bucket newest first, before tabs are built. The sort is stable; unknown timestamps sort last.
- In the TUI, `s` toggles newest/oldest first. `Model.applySort` re-sorts every tab by `Item.SortAt` and is re-applied when a refresh (`r`) delivers new tabs.

`parseTimestamp` accepts RFC 3339 with or without fractional seconds. An empty or invalid value becomes the zero time, which callers treat as oldest.
