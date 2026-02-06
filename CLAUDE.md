# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```sh
go build              # Build the extension
go test ./...         # Run all tests
go test -v ./...      # Run tests with verbose output
go test ./internal/ui # Run tests for a specific package
golangci-lint run     # Run linter
```

## Architecture

gh-own is a GitHub CLI extension that displays the user's owned PRs and issues in a terminal UI.

### Package Structure

- **cmd/** - Cobra command definitions (`root`, `pr`, `issue`)
- **internal/gh/** - GitHub API client wrapper using `cli/go-gh/v2`; generic `Search[T]` runs parallel API queries
- **internal/pullrequest/** - PR data types and search logic (groups by: created, assigned, review-requested, participated)
- **internal/issue/** - Issue data types and search logic (groups by: created, assigned, participated)
- **internal/ui/** - Bubbletea TUI with tabbed interface; `Model` manages tabs, `Item` represents list entries

### Data Flow

1. Commands call `gh.CurrentUser()` to get authenticated REST client and username
2. Domain packages (`pullrequest`, `issue`) call `gh.Search[T]()` with query conditions
3. Results are grouped and converted to UI tabs via `View()` methods in `*_ui.go` files
4. Bubbletea runs the interactive list; Enter opens selected item URL in browser
