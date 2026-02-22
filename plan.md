# Plan: PR Checkout via `g` keybinding

## Context

Users want to select a PR in gh-own's TUI and checkout the PR branch in the local repo managed by `ghq`. Currently `enter` opens the PR URL in a browser. The new `g` keybinding will find the repo via `ghq list --full-path -e <owner/repo>`, then run `gh pr checkout <number>` with WorkingDir set to that path. The TUI exits first so command output is visible in the terminal.

## Design

**Exit-then-execute pattern**: When `g` is pressed, the Model stores a "checkout request" (repoName + PR number) and quits the TUI. After `p.Run()` returns, the cmd layer checks for a checkout request and executes the external commands. This keeps the TUI simple and lets `gh pr checkout` output appear naturally in the terminal.

## Files to modify

- `internal/ui/ui.go` — Add `number` field to Item; add checkout request state to Model; add `g` keybinding
- `internal/ui/ui_test.go` — Tests for new behavior
- `internal/ui/styles.go` — Add `g` to help bar
- `internal/pr/ui.go` — Pass PR number to NewItem
- `internal/issue/ui.go` — Pass 0 for number (issues don't checkout)
- `internal/gh/pr.go` — (no changes needed — Number already available)
- `cmd/pr.go` — After Run(), check checkout request and execute ghq + gh pr checkout

## Step-by-step (TDD)

### Tidy 1: Add `number` field to `ui.Item`

Structural change — no behavior change.

- [x] **Test**: `NewItem` with number field — `item.Number()` returns the stored number
- Add `number int` field to `Item` struct
- Update `NewItem` signature: `NewItem(repoName, titleText, description, url string, number int) Item`
- Add `Number() int` accessor
- Add `RepoName() string` accessor (needed later by checkout handler)
- Update all `NewItem` call sites (`internal/pr/ui.go`, `internal/issue/ui.go`, test files)

### Tidy 2: Add checkout request state to Model

Structural change — no behavior change yet.

- [x] **Test**: `Model.CheckoutRequest()` returns `("", 0, false)` by default
- Add fields `checkoutRepo string`, `checkoutNumber int` to Model
- Add method `CheckoutRequest() (repo string, number int, ok bool)`

### Step 3: `g` keybinding sets checkout request and quits

- [ ] **Test**: Pressing `g` on a model with a selected item sets checkout request and returns tea.Quit
- [ ] **Test**: Pressing `g` during filtering is a no-op (consistent with enter/refresh behavior)
- [ ] **Test**: Pressing `g` with no selected item is a no-op

### Step 4: Add `g` to help bar

- [ ] **Test**: `helpView(list.Unfiltered)` contains "checkout"
- [ ] **Test**: `helpView(list.Filtering)` does NOT contain "checkout"

### Step 5: Execute checkout in cmd layer

- [ ] **Test**: `Checkout(repoDir, number)` runs `gh pr checkout <number>` with Dir set
- Create `internal/checkout/checkout.go` with:
  - `FindRepoDir(repoName string) (string, error)` — runs `ghq list --full-path -e <owner/repo>`
  - `Checkout(repoDir string, number int) error` — runs `gh pr checkout <number>` with Dir
- Wire into `cmd/pr.go`: after `p.Run()`, if `fm.CheckoutRequest()` returns ok, call FindRepoDir then Checkout

## Verification

```sh
go test ./...                    # All tests pass
go build                         # Builds successfully
gh own pr                        # TUI works, press g on a PR
ghq list --full-path -e <repo>   # Verify ghq finds the repo
```
