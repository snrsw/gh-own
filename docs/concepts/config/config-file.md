---
type: Configuration
title: Config file
description: The YAML file at `~/.config/gh-own/config.yaml` and the CLI flags that override or extend it.
resource: https://github.com/snrsw/gh-own/blob/main/internal/config/config.go
tags: [config, yaml, flags]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T15:16:46Z }
---

# Location

`$XDG_CONFIG_HOME/gh-own/config.yaml`, falling back to `~/.config/gh-own/config.yaml`. A missing file is not an error; it yields the zero `Config`, which means "all defaults".

# Shape

```yaml
pr:
  queries:            # map key -> search query; overrides or adds tabs
    needsAction: "is:pr is:open draft:false review:changes-requested {owner}"
    myBots: "is:pr is:open author:app/renovate"   # non-default key -> custom tab
  needsAction:
    conversation: true  # default; false disables conversation reading
issue:
  queries: {}
exclude:
  bots: false         # true adds the dependabot/renovate/github-actions preset
  authors: []         # extra logins, "renovate[bot]" or "app/renovate" spelling
```

| Field | Go type | Meaning |
|---|---|---|
| `pr.queries` / `issue.queries` | `map[string]string` | Merged over the defaults per key. See [PR tabs](/tabs/pr-tabs.md) and [issue tabs](/tabs/issue-tabs.md) for default keys. |
| `pr.needsAction.conversation` | `*bool` | `nil` means true; read through `ConversationEnabled()`. |
| `exclude.bots` | `bool` | Same effect as `--no-bots`. |
| `exclude.authors` | `[]string` | Unioned with `--exclude-author`. |

`pr.needsAction.conversation` is a pointer so that "unset" can mean true. **Read it only through `ConversationEnabled()`**; a plain bool would make an absent key mean false.

When `conversation` is false, two things change together:

- The PR search query requests the small activity fields instead of the conversation (see [conversation](/data/conversation.md)).
- `pr.NewGroupedPullRequests` skips [Needs action promotion](/rules/needs-action-promotion.md).

# Flags

Persistent flags on the root command (`cmd/root.go`):

| Flag | Effect |
|---|---|
| `--org <org>` | Adds `org:<org>` to every query |
| `--exclude-author <login>` | Repeatable; adds to `exclude.authors` |
| `--no-bots` | Same as `exclude.bots: true` |
| `--demo` | Uses fixture data, no API calls |
| `--debug` | Enables slog debug output |

How these become query text is described in [search pipeline](/pipelines/search.md).
