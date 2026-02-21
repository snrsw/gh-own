# UX Improvement Issues

## Navigation & Interaction

### 1. ~~No backward tab navigation~~ ✅ Resolved

~~`Tab` cycles forward through tabs, but there is no `Shift+Tab` or left/right arrow key to go back. Users must cycle through all tabs to return to a previous one.~~

**Impact:** Medium
**Status:** Resolved — `Shift+Tab` cycles backward through tabs with wrap-around.

### 2. ~~No loading indicator~~ ✅ Resolved

~~When the app starts, API queries run in parallel but the user sees nothing until data arrives. There is no spinner or "Loading..." feedback.~~

**Impact:** High
**Status:** Resolved in PR #39 — loading spinner wired up for both PR and Issue commands.

### 3. No keyboard shortcut to jump to a specific tab

Users cannot press `1`/`2`/`3`/`4` to jump directly to a tab.

**Impact:** Low

### 4. ~~No way to refresh data~~ ✅ Resolved

~~Once loaded, data is static. There is no `r` key to re-fetch without restarting the command.~~

**Impact:** Medium
**Status:** Resolved — pressing `r` re-fetches data with loading spinner, ignored during filtering and loading states.

## Information Display

### 5. No empty-state message per tab

When a tab has 0 items, the list is just blank. A message like "No review-requested PRs" would be more helpful.

**Impact:** Medium

### 6. PR labels/tags not shown

Labels (e.g., `bug`, `enhancement`) are not fetched or displayed, which limits at-a-glance triage.

**Impact:** Low

### 7. ~~No review status indicator for PRs~~ ✅ Resolved

~~Beyond CI status, there is no indication of review state (approved, changes requested, pending review) on the list item itself.~~

**Impact:** Medium
**Status:** Resolved in PR #41 — review status indicator added to PR list items.

### 8. Comment count not shown

The number of comments is not displayed, which is useful for gauging activity level.

**Impact:** Low

### 9. Draft PRs lack visual distinction

Draft PRs are only marked in the number prefix but do not have a distinct visual style (e.g., dimmed title or a `[Draft]` badge) making them hard to spot.

**Impact:** Low

## Command & Configuration

### 10. No unified PR+Issue view

`pr` and `issue` are separate subcommands. A unified view showing both PRs and issues together would reduce context-switching.

**Impact:** Medium

### 11. No `--repo` or `--org` filter flag

Users cannot scope results to a specific repo or org without modifying the code.

**Impact:** Low

### 12. No sort options

Items appear in API return order. Users cannot sort by updated date, created date, or activity recency.

**Impact:** Medium

## Visual Polish

### 13. No total count summary

The tab bar does not indicate total count across all tabs. A summary like "12 open PRs" at the top would provide quick context.

**Impact:** Low

### 14. ~~Ambiguous "updated" fallback~~ ⚠️ Partially Resolved

~~"updated 2w ago" does not tell you what was updated.~~ The latest activity feature now shows specific actions (commented/approved/pushed) in PR and issue lists, but items without tracked activity still fall back to the generic "updated" text.

**Impact:** Low
**Status:** Partially resolved in PR #38 — latest activity (commented/approved/pushed) shown in lists.

### 15. Static help bar

The help footer always shows the same keys. Context-sensitive help (e.g., showing filter-mode keys when filtering) would be clearer.

**Impact:** Low
