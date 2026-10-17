---
type: Rule
title: Needs action promotion
description: Moves PRs whose conversation is waiting on the user into the Needs action tab and out of the other default tabs.
resource: https://github.com/snrsw/gh-own/blob/main/internal/pr/needsaction.go
tags: [pr, needs-action, tabs]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Result

**Needs action = PRs matched by the `needsAction` query (changes requested) ∪ PRs whose [attention](/rules/attention.md) is non-zero.** This is the only place where a tab holds more than its query matches.

# Algorithm

`promoteNeedsAction(r, login)` is called from `pr.NewGroupedPullRequests` when conversation reading is on (see [config file](/config/config-file.md)):

1. Annotate every existing Needs action PR with its attention (`owned = true`) and keep all of them, even those with no attention.
2. Visit the other buckets in this order, computing attention for each PR not seen yet:

   | Order | Bucket | `owned` | Promoted PR removed from bucket? |
   |---|---|---|---|
   | 1 | Drafts | true | yes |
   | 2 | Ready to merge | true | yes |
   | 3 | Waiting | true | yes |
   | 4 | Participated | false | yes |
   | 5 | Review Requested | false | **no** |
   | — | Custom tabs | — | not visited |

3. A PR with non-zero attention is appended to Needs action once. URL dedup ensures a PR returned by several queries is judged once, in the first bucket that contains it.

Review Requested keeps promoted PRs because the review is still due, whatever else happened.

# Disabling

- `login == ""` returns the input unchanged. Demo mode relies on this (see [search pipeline](/pipelines/search.md)).
- `pr.needsAction.conversation: false` skips the call entirely.

# Effect on display

A promoted PR's description line and sort time come from `Attention`, not `LatestActivity`. See [activity line and sort order](/ui/activity-and-sort.md).
