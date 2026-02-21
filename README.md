# gh-own

[![CI](https://github.com/snrsw/gh-own/actions/workflows/ci.yml/badge.svg)](https://github.com/snrsw/gh-own/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/snrsw/gh-own)](https://goreportcard.com/report/github.com/snrsw/gh-own)
[![GitHub release](https://img.shields.io/github/v/release/snrsw/gh-own)](https://github.com/snrsw/gh-own/releases/latest)
[![Go version](https://img.shields.io/github/go-mod/go-version/snrsw/gh-own)](go.mod)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

GitHub CLI extension to list your owned PRs and issues across repositories unlike `gh pr list` and `gh issue list`, which are repo-scoped.

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
- Displays CI status (✓ / ✗ / ●) and review decision (✔ / ⊘ / ◇) for each PR
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

## Requirements

- [GitHub CLI](https://cli.github.com/) installed and authenticated
