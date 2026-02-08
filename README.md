# gh-own

GitHub CLI extension to list your owned pull requests and issues across repositories unlike `gh pr list` and `gh issue list`, which are repo-scoped.

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
