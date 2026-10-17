---
type: DataModel
title: Conversation
description: The tail of a PR's discussion (description, comments, reviews, review threads, commits) fetched with each search result.
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/conversation.go
tags: [pr, graphql, conversation]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Purpose

`gh.Conversation` is the input of the [attention rule](/rules/attention.md). It is fetched inside the same GraphQL search query as the PR, so it costs no extra requests, only a larger response.

# Fields

| Field | GraphQL source | Fetched |
|---|---|---|
| `Author`, `CreatedAt`, `Body` | PR `author.login`, `createdAt`, `body` | — |
| `Comments []Comment` | `comments(last: 10)` | last 10 |
| `Reviews []Review` | `reviews(last: 10)` | last 10, PENDING dropped |
| `Threads []ReviewThread` | `reviewThreads(last: 10)` → `comments(last: 10)` | 10 most recently **opened** threads, last 10 comments each |
| `Commits []Commit` | `commits(last: 5)` | last 5 |

Every list is chronological, oldest first. The page sizes are the constants `conversationPageSize` (10) and `commitsPageSize` (5). Both `prConversationFields` (the query) and [truncation horizon](/rules/truncation-horizon.md) are built from them. **Change page sizes only through these constants**, or the query and the horizon will disagree.

# Parsing rules

These live in `prSearchRawNode.conversation()` in `internal/gh/pr.go`:

- **Bot detection** uses the GraphQL `__typename`: `Bot` sets `Comment.IsBot`. In GraphQL, an App's login has no `[bot]` suffix, so matching on the login string would fail.
- **PENDING reviews are dropped.** They are unsubmitted, and only their author can see them.
- **Thread comment time is `publishedAt`, falling back to `createdAt`.** A comment drafted inside a review is created when typed but becomes visible when the review is submitted. Issue comments use `createdAt`, since they are published on creation.
- **Commit login** is `commit.author.user.login`. It is empty when the git author does not map to a GitHub user.
- **A null author** (a deleted account) gives an empty login. The attention rule ignores such events.

# Conversation off

When `pr.needsAction.conversation` is false ([config file](/config/config-file.md)), the query uses `prActivityFields` instead: only `comments(last: 1)`, `reviews(last: 1)`, `commits(last: 1)`, with no bodies and no threads. The `Conversation` then holds just enough for the [activity line](/ui/activity-and-sort.md).

The CI status always comes from the last fetched commit's `statusCheckRollup.state`, whichever field set is used. See [status indicators](/ui/status-indicators.md).

# Known blind spot

GitHub orders `reviewThreads` by when a thread was opened, not by its last reply. On a PR with more than 10 threads, a reply in an older thread is not fetched.
