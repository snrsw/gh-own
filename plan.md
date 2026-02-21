# Plan: Add Loading Indicator

Show a spinner with "Loading..." while API data is being fetched, instead of a blank terminal.

## Problem

Currently, all data fetching happens synchronously in `cmd/pr.go` and `cmd/issue.go` **before** Bubbletea starts. The user sees nothing until all API queries complete (typically 2-5 seconds).

## Approach

Introduce a **loading state** in `ui.Model` that displays a spinner. The Bubbletea program starts immediately with this state, and data fetching runs as a `tea.Cmd`. When results arrive, the model transitions to the normal tabbed view.

### Key Design Decisions

1. **Use `charmbracelet/bubbles/spinner`** — standard Bubbletea spinner component, already a transitive dependency
2. **Move data fetching into a `tea.Cmd`** — the command handlers pass a fetch function to the UI, which runs it asynchronously
3. **Two-phase Model** — `loading: true` shows spinner, `loading: false` shows tabs (current behavior)

## UI Design

Loading state:
```
  ⣾ Loading...
```

After data arrives: current tabbed UI (no change).

Error state: display error message and quit.

## Implementation (TDD)

### Test 1: Model starts in loading state when created via NewLoadingModel
- [x] `NewLoadingModel()` returns a Model with `loading == true`
- [x] `View()` in loading state renders spinner text (contains "Loading")
- [x] `Init()` returns a non-nil `tea.Cmd` (the spinner tick)

### Test 2: Spinner animates on tick messages
- [x] In loading state, `Update(spinner.TickMsg)` returns a command (keeps spinner alive)
- [x] In loaded state, `Update(spinner.TickMsg)` is ignored (no command)

### Test 3: TabsMsg transitions from loading to loaded state
- [x] Define `TabsMsg` as `[]Tab`
- [x] `Update(TabsMsg)` sets `loading = false` and populates tabs
- [x] After receiving `TabsMsg`, `View()` renders tabs (not spinner)

### Test 4: ErrMsg displays error and quits
- [x] Define `ErrMsg` as `struct{ Err error }`
- [x] `Update(ErrMsg)` returns `tea.Quit` command

### Test 5: FetchCmd wraps a function into a tea.Cmd returning TabsMsg or ErrMsg
- [x] `FetchCmd(fn func() ([]Tab, error))` returns a `tea.Cmd`
- [x] When `fn` succeeds, the command returns `TabsMsg`
- [x] When `fn` fails, the command returns `ErrMsg`

### Test 6: Window resize during loading state sets dimensions
- [x] In loading state, `Update(WindowSizeMsg)` stores width/height
- [x] After transition to loaded, tabs have correct sizes

### Test 7: Wire up PR command to use loading model
- [x] `cmd/pr.go` calls `ui.NewLoadingModel()` and passes `FetchCmd` with PR fetching logic
- [x] `pr.BuildTabs()` extracts tab-building from `pr.View()` so it can be called inside `FetchCmd`
- [ ] Manual verification: `go build && gh own pr` shows spinner then results

### Test 8: Wire up Issue command to use loading model
- [x] `cmd/issue.go` calls `ui.NewLoadingModel()` and passes `FetchCmd` with issue fetching logic
- [x] `issue.BuildTabs()` extracts tab-building from `issue.View()` so it can be called inside `FetchCmd`
- [ ] Manual verification: `go build && gh own issue` shows spinner then results

## Files to Modify/Create

| File | Action |
|------|--------|
| `internal/ui/ui.go` | Modify - Add loading state, `NewLoadingModel`, `TabsMsg`, `ErrMsg`, `FetchCmd`, spinner |
| `internal/ui/ui_test.go` | Modify - Add tests for loading state, transitions, fetch command |
| `internal/pr/ui.go` | Modify - Extract tab-building into `BuildTabs`, remove `tea.NewProgram` call |
| `internal/issue/issue.go` or new `internal/issue/ui.go` | Modify - Extract tab-building, remove `tea.NewProgram` call |
| `cmd/pr.go` | Modify - Start TUI immediately with loading model, pass fetch function |
| `cmd/issue.go` | Modify - Same as pr.go |

## Verification

1. `go test ./...` after each test
2. `go build` to verify compilation
3. `gh own pr` — spinner appears, then PR list loads
4. `gh own issue` — spinner appears, then issue list loads
5. Test with slow network to confirm spinner is visible
