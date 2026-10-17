---
type: Rule
title: Truncation horizon
description: The cutoff time before which conversation events are ignored, because the user's answer may lie outside the fetched tail.
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/conversation.go
tags: [pr, conversation, correctness]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Problem

The [conversation](/data/conversation.md) holds only the last N entries of each list. Example: a PR has 30 comments, and the user replied at comment 12. Only comments 21–30 are fetched, so comment 12 is not visible. A mention at comment 21 might then already be answered, but nothing in the fetched data shows it.

# Rule

`Conversation.horizon()` returns a time `h`, and the [attention rule](/rules/attention.md) ignores every event before `h`.

- A list is **full** when it holds as many entries as its page size: 10 for comments and reviews, 5 for commits.
- For each full list, take the time of its oldest fetched entry.
- `h` is the newest of those times. The zero time means no list is full.

# Deliberate choices

| Choice | Reason |
|---|---|
| Review threads are excluded | Threads are ordered by creation, not by activity, so a full thread page says nothing about which replies are hidden. |
| Reviews are counted after PENDING ones are dropped | A page holding a pending review then looks not full. The horizon becomes more permissive, which is the harmless direction. |
| Old events are dropped, not kept | On a very busy PR, an old unanswered mention can drop out of Needs action. This was chosen over keeping already-answered items there forever. |

# Coupling

The horizon reads `conversationPageSize` and `commitsPageSize`, the same constants the GraphQL query is built from. If the query fetches more or fewer entries than these constants say, the horizon is wrong.
