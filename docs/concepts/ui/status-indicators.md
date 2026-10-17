---
type: Enum
title: Status indicators
description: The CI and review status symbols appended to each PR title, and the GitHub GraphQL values they map from.
resource: https://github.com/snrsw/gh-own/blob/main/internal/cistatus/cistatus.go
tags: [ui, pr, ci, review]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Placement

A PR title renders as `#123 Title <review> <ci>` (`pullRequest.toItem` in `internal/pr/ui.go`). The review symbol is omitted when it is empty. Draft PRs show `[DRAFT] #123` in gray instead of `#123` in blue (`#0969DA`).

# CI status

Source: `statusCheckRollup.state` of the last fetched commit (see [conversation](/data/conversation.md)). Package: `internal/cistatus`.

| GraphQL state | `CIStatus` | Symbol | Color |
|---|---|---|---|
| `SUCCESS` | `CIStatusSuccess` | ✓ | `#1A7F37` |
| `FAILURE`, `ERROR` | `CIStatusFailure` | ✗ | `#CF222E` |
| `PENDING`, `EXPECTED` | `CIStatusPending` | ● | `#9A6700` |
| anything else, or no rollup | `CIStatusNone` | - | `#6E7781` |

# Review status

Source: the PR's `reviewDecision`. Package: `internal/reviewstatus`.

| GraphQL decision | `ReviewStatus` | Symbol | Color |
|---|---|---|---|
| `APPROVED` | `ReviewStatusApproved` | ✔ | `#1A7F37` |
| `CHANGES_REQUESTED` | `ReviewStatusChangesRequested` | ⊘ | `#CF222E` |
| `REVIEW_REQUIRED` | `ReviewStatusReviewRequired` | ◇ | `#9A6700` |
| anything else | `ReviewStatusNone` | (none) | — |

Colors are GitHub's Primer palette, the same for both enums. Issues show neither indicator.
