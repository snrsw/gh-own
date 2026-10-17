---
type: TabSet
title: Issue tabs
description: The three default tabs of `gh-own issue`; issues have no conversation reading and no tab promotion.
resource: https://github.com/snrsw/gh-own/blob/main/internal/gh/issue.go
tags: [issue, tabs, query-keys]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Tabs

| Tab | Bucket field | Query key | Default query |
|---|---|---|---|
| Created | `Created` | `created` | `is:issue is:open author:{user}` |
| Assigned | `Assigned` | `assigned` | `is:issue is:open assignee:{user}` |
| Participated | `Participated` | `participatedUser` + `participatedTeam<N>` | `is:issue is:open involves:{user} -author:{user} -assignee:{user}` |
| Custom (one per extra key) | `Custom[key]` | any non-default key | from config |

Key routing works as for [PR tabs](/tabs/pr-tabs.md): the `participated` prefix collects both the user query and the team queries.

# Differences from PR tabs

- **Every issue tab is exactly its query.** No conversation is fetched, so the [attention rule](/rules/attention.md) and [Needs action promotion](/rules/needs-action-promotion.md) do not apply.
- The GraphQL query fetches only `comments(last: 1)`. The activity line can only report "commented"; there is no review or push.
- Issues carry no CI or review status. See [status indicators](/ui/status-indicators.md).
