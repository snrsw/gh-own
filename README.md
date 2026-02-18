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

## Installation

```sh
gh extension install snrsw/gh-own
```

## Usage

```sh
gh own [command]
```

### Commands

| Command | Description |
|---------|-------------|
| `gh own` | List your pull requests (default) |
| `gh own pr` | List your pull requests |
| `gh own issue` | List your issues |

### Examples

```sh
# List your pull requests (default behavior)
gh own

# Explicitly list pull requests
gh own pr

# List your issues
gh own issue
```

## Requirements

- [GitHub CLI](https://cli.github.com/) installed and authenticated
