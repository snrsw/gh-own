# gh-own

[![CI](https://github.com/snrsw/gh-own/actions/workflows/ci.yml/badge.svg)](https://github.com/snrsw/gh-own/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/snrsw/gh-own)](https://goreportcard.com/report/github.com/snrsw/gh-own)
[![GitHub release](https://img.shields.io/github/v/release/snrsw/gh-own)](https://github.com/snrsw/gh-own/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/snrsw/gh-own)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

GitHub CLI extension to list your owned PRs and issues across repositories unlike `gh pr list` and `gh issue list`, which are repo-scoped.

![demo](demo/demo.gif)

Key features:

- List your pull requests across all repositories grouped by what to do next:
  - Needs action — non-draft PRs where reviewers requested changes
  - Ready to merge — non-draft PRs that are approved
  - Requested your review (including teams)
  - Waiting for review or checks — non-draft PRs not yet approved and without changes requested
  - Drafts — your open draft PRs
  - You have participated in mentioned or commented (including teams)

  The state-based tabs (Needs action, Ready to merge, Waiting, Drafts) cover PRs you authored or are assigned to.
- List your issues across all repositories grouped into:
  - Created by you
  - Assigned to you
  - You have participated in mentioned or commented (including teams)
- Displays CI status and review decision for each PR (see [Symbol legend](#symbol-legend))
- Shows latest activity (who commented, reviewed, or pushed and when)
- Includes draft PR indication
- Fetches results for all teams you belong to, merged and deduplicated with your personal results
- Team slugs are cached for 6 hours to avoid repeated API calls
- Filter results to a single GitHub organization with `--org`
- Hide bot-authored pull requests and issues with `--no-bots` or `--exclude-author`

## Installation

```sh
gh extension install snrsw/gh-own
```

## Usage

```sh
gh own [command] [flags]
```

### Commands

| Command | Description |
|---------|-------------|
| `gh own` | List your pull requests (default) |
| `gh own pr` | List your pull requests |
| `gh own issue` | List your issues |

### Flags

| Flag | Description |
|------|-------------|
| `--org <name>` | Filter results to a single GitHub organization |
| `--exclude-author <login>` | Hide items written by this author; repeatable, or comma-separated |
| `--no-bots` | Hide items written by Dependabot, Renovate, and GitHub Actions |
| `--demo` | Use built-in demo data without calling the GitHub API (useful for screenshots and recordings) |
| `--debug` | Enable debug logging to stderr (includes timing instrumentation) |

### Examples

```sh
# List your pull requests (default behavior)
gh own

# Explicitly list pull requests
gh own pr

# List your issues
gh own issue

# Filter to a specific organization
gh own --org my-org

# Hide the usual bots
gh own --no-bots

# Hide a specific author
gh own --exclude-author 'my-release-bot'

# Enable debug logging
gh own --debug

```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Switch between tabs |
| `enter` | Open selected item in browser |
| `r` | Refresh data |
| `s` | Toggle sort order (newest / oldest first) |
| `/` | Filter items in current tab |
| `ctrl+c` | Quit |

### Sort order

Every tab lists items newest first, using the same timestamp shown on each item's
description line — the latest activity (comment, review, or push) when there is
one, and the item's last-updated time otherwise. Press `s` to flip to oldest
first; the choice applies to all tabs, keeps any filter you have applied, and
survives a refresh.

Searches are also issued with `sort:updated-desc` so that the page GitHub returns
is the most recently updated items rather than the most relevant ones. If one of
your [configured queries](#configuration) already specifies its own `sort:`
qualifier, that one is used instead. The team searches behind the Participated
tab are not configurable and always use `sort:updated-desc`.

## Symbol legend

### CI status

| Symbol | Meaning |
|--------|---------|
| `✓` | CI passed |
| `✗` | CI failed |
| `●` | CI pending |
| `-` | No CI status |

### Review decision

| Symbol | Meaning |
|--------|---------|
| `✔` | Approved |
| `⊘` | Changes requested |
| `◇` | Review required |

## Configuration

You can customize the search queries used for each tab by creating a config file at `$XDG_CONFIG_HOME/gh-own/config.yaml` (defaults to `~/.config/gh-own/config.yaml`).

Use the `{user}` placeholder to reference the authenticated GitHub username.

```yaml
pr:
  queries:
    readyToMerge: "is:pr is:open draft:false review:approved {owner} label:team-a"
    review_requested: "is:pr is:open review-requested:{user} label:urgent"
issue:
  queries:
    participated: "is:issue is:open involves:{user}"
```

Any query you specify overrides the default for that tab. Tabs you don't specify keep their defaults. If no config file exists, the extension behaves exactly as before.

The four state-based PR tabs (Drafts, Needs action, Ready to merge, Waiting) use the `{owner}` placeholder, which expands to both `author:{user}` and `assignee:{user}` — because GitHub search cannot match `author:` or `assignee:` in a single query. The two searches merge into one tab, so a single `{owner}` query covers both the PRs you authored and the ones assigned to you.

> **Migration note:** the `pr` command previously had `created` and `assigned` tabs. They have been replaced by the four state-based tabs above. If your config still overrides `created` or `assigned` under `pr.queries`, those keys now render as extra [custom tabs](#custom-tabs) rather than replacing a default — rename them to the new keys to restore the intended behavior.

### Custom tabs

You can add custom tabs by defining queries with non-default keys. Custom tabs appear after the default tabs, sorted alphabetically. The key name is used as the tab title (hyphens become spaces, each word capitalized).

```yaml
pr:
  queries:
    needs-triage: "is:pr is:open label:needs-triage"
    team-review: "is:pr is:open team-review-requested:my-org/my-team"
issue:
  queries:
    bugs: "is:issue is:open label:bug"
```

This adds tabs named "Needs Triage", "Team Review", and "Bugs" respectively.

### Excluding authors

Bots such as Renovate and Dependabot can bury the pull requests you actually
need to look at. The `exclude` section drops their items from every tab of both
commands:

```yaml
exclude:
  bots: true
  authors:
    - "renovate[bot]"
    - "my-release-bot"
```

| Key | Meaning |
|-----|---------|
| `bots` | Turn on the built-in preset: `app/dependabot`, `app/renovate`, `app/github-actions` |
| `authors` | Additional logins to hide |

The `--no-bots` and `--exclude-author` flags do the same thing for a single run.
The config file, the flags, and the preset combine — they do not override each
other — so `exclude.bots: true` in your config plus `--exclude-author noisy-bot`
on the command line hides all four accounts.

Both spellings GitHub accepts work: the API login `renovate[bot]` and the search
qualifier `app/renovate` refer to the same account, and listing it twice in
different spellings is harmless.

Exclusion happens inside the search query, not after the results come back. This
matters because each query returns a fixed-size page: without exclusion, fifty
Renovate pull requests can fill that page and push the ones you care about out
of view entirely.

A query of your own that already names an excluded author keeps its own meaning,
the same way a query with its own `sort:` keeps that sort. So a custom tab built
around `author:app/renovate` still works while `exclude.bots` is on.

> If you list an author that does not exist, GitHub rejects the whole search with
> a "Validation Failed" error naming the problem. Check the spelling of the login.

### Default queries

The built-in defaults are equivalent to the following config:

```yaml
pr:
  queries:
    drafts: "is:pr is:open draft:true {owner}"
    needsAction: "is:pr is:open draft:false review:changes-requested {owner}"
    readyToMerge: "is:pr is:open draft:false review:approved {owner}"
    waiting: "is:pr is:open draft:false -review:approved -review:changes-requested {owner}"
    review_requested: "is:pr is:open review-requested:{user}"
    participated: "is:pr is:open involves:{user} -author:{user} -assignee:{user} -review-requested:{user}"
issue:
  queries:
    created: "is:issue is:open author:{user}"
    assigned: "is:issue is:open assignee:{user}"
    participated: "is:issue is:open involves:{user} -author:{user} -assignee:{user}"
```

### Available keys

| Command | Keys |
|---------|------|
| `pr` | `drafts`, `needsAction`, `readyToMerge`, `waiting`, `review_requested`, `participated` |
| `issue` | `created`, `assigned`, `participated` |

## Requirements

- [GitHub CLI](https://cli.github.com/) installed and authenticated
