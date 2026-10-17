---
type: Rule
title: Attention
description: Decides whether a PR's conversation is waiting on the user, and why (mentioned you / replied to you / commented).
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/attention.go
tags: [pr, needs-action, conversation, mentions]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Signature

`func (c Conversation) Attention(login string, owned bool) Attention`

It returns `Attention{Reason, Login, At}`. The zero value means nothing is waiting. An empty `login` always returns the zero value.

`owned` is true for PRs from the owned buckets (Drafts, Ready to merge, Waiting, Needs action; see [PR tabs](/tabs/pr-tabs.md)).

# Reasons

From strongest to weakest:

| Reason constant | Text | Rank | Applies to | Triggered by |
|---|---|---|---|---|
| `AttentionMentioned` | `mentioned you` | 3 | any PR | @-mention of the user in the PR body, a comment, a review body, or a thread comment |
| `AttentionReplied` | `replied to you` | 2 | any PR | a comment by someone else after the user's comment, in an **unresolved** thread |
| `AttentionCommented` | `commented` | 1 | owned PRs only | a comment, a thread comment, or a `COMMENTED` / `CHANGES_REQUESTED` review, by a **non-bot** |

Selection: the highest rank wins. Among events with equal rank, the most recent wins.

# Events that never count

`attentionFinder.add` drops an event when any of these holds:

- Its author is the user.
- Its author login is empty (a deleted account).
- Its time is unparsable, or not after the user's last activity.
- Its time is before the [truncation horizon](/rules/truncation-horizon.md).

The PR body's mention is also skipped when the user is the PR author.

**The user's last activity** is computed by `Conversation.lastActivityBy`. It is the latest of: opening the PR, commenting, reviewing, commenting in a thread, or authoring a commit. Any commit counts, including a base-branch merge or "Update branch". Everything before the last activity is treated as answered.

Two exclusions are deliberate:

- Bot comments never count as `commented`, because CI and coverage bots comment on nearly every push. Bot mentions and bot thread replies still count.
- `APPROVED` reviews never count as `commented`, because approval is what the Ready to merge tab is for. A mention inside an approving review still counts.

# Mention matching

`mentions(body, login)` follows GitHub's notification behavior:

- Case-insensitive.
- Whole login only. `@alice` matches; `@alice-bot`, `bob@alice`, and the team mention `@alice/core` do not. Login characters are `[A-Za-z0-9_-]`.
- Text in fenced code blocks, inline code spans, and `>` quoted lines is removed first (`stripQuotedAndCode`).
- Inline code follows CommonMark: a backtick run closes only at the next run of the same length. An unclosed backtick is literal text.

# Consumer

[Needs action promotion](/rules/needs-action-promotion.md) calls this and stores the result in `PRSearchNode.Attention`. The tests in `internal/gh/attention_test.go` are the executable spec; add a case there for any rule change.
