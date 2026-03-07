# gh-own

[![CI](https://github.com/snrsw/gh-own/actions/workflows/ci.yml/badge.svg)](https://github.com/snrsw/gh-own/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/snrsw/gh-own)](https://goreportcard.com/report/github.com/snrsw/gh-own)
[![GitHub release](https://img.shields.io/github/v/release/snrsw/gh-own)](https://github.com/snrsw/gh-own/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/snrsw/gh-own)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

GitHub CLI extension to list your owned PRs and issues across repositories unlike `gh pr list` and `gh issue list`, which are repo-scoped.

![demo](demo/demo.gif)

Key features:

- List your pull requests across all repositories grouped into:
  - Created by you
  - Assigned to you
  - Requested your review (including teams)
  - You have participated in mentioned or commented (including teams)
- List your issues across all repositories grouped into:
  - Created by you
  - Assigned to you
  - You have participated in mentioned or commented (including teams)
- Displays CI status and review decision for each PR (see [Symbol legend](#symbol-legend))
- Shows latest activity (who commented, reviewed, or pushed and when)
- Includes draft PR indication
- Fetches results for all teams you belong to, merged and deduplicated with your personal results
- Team slugs are cached for 6 hours to avoid repeated API calls

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
| `--debug` | Enable debug logging to stderr (includes timing instrumentation) |

### Examples

```sh
# List your pull requests (default behavior)
gh own

# Explicitly list pull requests
gh own pr

# List your issues
gh own issue

# Enable debug logging
gh own --debug

```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `tab` / `shift+tab` | Switch between tabs |
| `enter` | Open selected item in browser |
| `r` | Refresh data |
| `/` | Filter items in current tab |
| `ctrl+c` | Quit |

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
    created: "is:pr is:open author:{user} label:team-a"
    review_requested: "is:pr is:open review-requested:{user} label:urgent"
issue:
  queries:
    participated: "is:issue is:open involves:{user}"
```

Any query you specify overrides the default for that tab. Tabs you don't specify keep their defaults. If no config file exists, the extension behaves exactly as before.

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

### Default queries

The built-in defaults are equivalent to the following config:

```yaml
pr:
  queries:
    created: "is:pr is:open author:{user}"
    assigned: "is:pr is:open assignee:{user}"
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
| `pr` | `created`, `assigned`, `review_requested`, `participated` |
| `issue` | `created`, `assigned`, `participated` |

## Requirements

- [GitHub CLI](https://cli.github.com/) installed and authenticated
