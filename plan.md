# Plan: Refresh Data (Issue #4)

## Goal

Add `r` key binding to re-fetch API data without restarting the command. Once loaded, data is currently static — users must quit and re-run to see updates.

## Approach

The `Model` already stores `fetchCmd` for the initial load. When the user presses `r`:

1. Transition back to loading state (`m.loading = true`)
2. Re-dispatch `m.fetchCmd` along with a spinner tick
3. The existing `TabsMsg` / `ErrMsg` handling completes the cycle

Key details:
- `r` is ignored during loading state (prevent double-fetch)
- `r` is ignored during filter mode (let the user type freely)
- The help bar gets a new `r → refresh` entry
- Active tab resets to 0 on refresh (same as initial load behavior)

## Files Changed

- `internal/ui/ui.go` — add `r` key handler, `handleRefresh()` method
- `internal/ui/styles.go` — add `r` / `refresh` help entry
- `internal/ui/ui_test.go` — new tests

## Tests (TDD order)

### 1. ✅ `r` key triggers refresh command
File: `internal/ui/ui_test.go`
Create a loaded `Model` with `fetchCmd` set. Press `r`. Assert `loading` becomes `true` and a non-nil command is returned.

### 2. ✅ `r` key is ignored during loading
File: `internal/ui/ui_test.go`
Create a `NewLoadingModel`. Press `r`. Assert no command is returned and state is unchanged.

### 3. ✅ `r` key is ignored during filter mode
File: `internal/ui/ui_test.go`
Create a loaded `Model`, activate filter mode, press `r`. Assert `loading` remains `false` (the key should pass through to the list filter).

### 4. ✅ Help bar includes refresh shortcut
File: `internal/ui/ui_test.go`
Assert `helpView()` output contains `"r"` and `"refresh"`.

### 5. ✅ Refresh followed by TabsMsg completes cycle
File: `internal/ui/ui_test.go`
Press `r` to enter loading, then send `TabsMsg` with new tabs. Assert `loading` is `false` and tabs are updated.
