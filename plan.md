# Plan: Add Latest Activity to PR and Issue Lists

## Context

`gh-own` currently shows each PR/issue with: `#N opened on DATE by AUTHOR, updated AGO`.
This tells you *when* but not *what happened*. Adding "latest activity" shows the most
recent meaningful event: who commented, who approved/requested-changes, or who pushed.

Desired output examples:
- `#42 opened on 2024-03-10 by alice · approved by bob 2h ago`
- `#7 opened on 2024-01-15 by carol · commented by dave 5m ago`
- `#99 opened on 2024-02-01 by eve · changes requested by frank 3d ago`
- `#12 opened on 2024-03-20 by grace · pushed by henry 1h ago`
- `#5 opened on 2024-03-18 by ivan, updated 2d ago`  ← fallback when no activity

## Approach

Extend the existing GraphQL queries to fetch:
- `comments(last: 1)` — last comment (PRs and Issues)
- `reviews(last: 1)` — last review with state (PRs only)
- `commits(last: 1)` — already fetched; extend to include commit author + date (PRs only)

Pick the most-recent event across those sources and display it in the description line.
If none available, fall back to the existing "updated X ago" format.

No new dependencies. No changes to `cmd/` or `main.go`.

---

## Steps

### Step 1: `LatestActivity` type and picker in `internal/gh`

**Files:** `internal/gh/activity.go`, `internal/gh/activity_test.go`

`LatestActivity` stores the resolved "most recent event":
```go
type LatestActivity struct {
    Kind  string // "commented", "approved", "changes requested", "dismissed", "pushed", ""
    Login string // actor username; empty means no activity
    At    string // RFC3339 timestamp
}
```

`NewLatestActivity(commentLogin, commentAt, reviewLogin, reviewAt, reviewState, pushLogin, pushAt string) LatestActivity`
selects the most recent non-empty event across comment, review, and push.

Tests (`activity_test.go`):
- [x] `TestNewLatestActivity_NoActivity` — all empty → `LatestActivity{}` with empty Login
- [x] `TestNewLatestActivity_CommentOnly` — only comment → Kind="commented", Login=comment author
- [x] `TestNewLatestActivity_ReviewApproved` — approved review → Kind="approved"
- [x] `TestNewLatestActivity_ReviewChangesRequested` — → Kind="changes requested"
- [x] `TestNewLatestActivity_ReviewDismissed` — → Kind="dismissed"
- [x] `TestNewLatestActivity_PushOnly` — only push → Kind="pushed"
- [x] `TestNewLatestActivity_CommentMoreRecentThanReview` — picks comment
- [x] `TestNewLatestActivity_ReviewMoreRecentThanComment` — picks review

---

### Step 2: Extend PR GraphQL query + parse into `PRSearchNode.LatestActivity`

**Files:** `internal/gh/pr.go`, `internal/gh/pr_test.go`

Extend `prSearchQuery` to add inside `... on PullRequest`:
```graphql
comments(last: 1) {
    nodes { author { login } createdAt }
}
reviews(last: 1) {
    nodes { author { login } submittedAt state }
}
```
Also extend existing `commits(last: 1)` node:
```graphql
commits(last: 1) {
    nodes {
        commit {
            statusCheckRollup { state }
            committedDate
            author { user { login } }
        }
    }
}
```

Add corresponding fields to `prSearchRawNode` and populate `PRSearchNode.LatestActivity`
via `NewLatestActivity` in `parsePRSearchNodes`.

Tests (`pr_test.go`):
- [x] `TestParsePRSearchNodes_WithComment` — raw node has comment → `LatestActivity.Kind == "commented"`
- [x] `TestParsePRSearchNodes_WithApprovedReview` — raw node has approved review → `Kind == "approved"`
- [x] `TestParsePRSearchNodes_WithPush` — raw node has commit author → `Kind == "pushed"`
- [x] `TestParsePRSearchNodes_NoActivity` — no comment/review/push author → `LatestActivity.Login == ""`

---

### Step 3: Propagate to `pullRequest` and update `toItem()` description

**Files:** `internal/pr/pr.go`, `internal/pr/ui.go`, `internal/pr/pr_test.go`

Add `LatestActivity gh.LatestActivity` to `pullRequest`. Copy it in `fromGraphQL`.

In `toItem()` description:
- If `LatestActivity.Login != ""`: `"#N opened on DATE by AUTHOR · KIND by LOGIN AGO"`
- Else: fall back to `"#N opened on DATE by AUTHOR, updated AGO"` (existing format)

Tests (`pr_test.go`):
- [x] `TestPullRequest_ToItem_WithActivity` — description contains `"· approved by bob"`
- [x] `TestPullRequest_ToItem_NoActivity` — description still contains `"updated"` (fallback)
- [x] `TestFromGraphQL_PropagatesLatestActivity` — `fromGraphQL` copies `LatestActivity`

---

### Step 4: Extend issue GraphQL query + parse into `IssueSearchNode.LatestActivity`

**Files:** `internal/gh/issue.go`, `internal/gh/issue_test.go`

Extend `issueSearchQuery` to add inside `... on Issue`:
```graphql
comments(last: 1) {
    nodes { author { login } createdAt }
}
```

Add fields to `issueSearchRawNode` and populate `IssueSearchNode.LatestActivity`
(only comment; no reviews or commits for issues).

Tests (`issue_test.go`):
- [x] `TestParseIssueSearchNodes_WithComment` — raw node has comment → `LatestActivity.Kind == "commented"`
- [x] `TestParseIssueSearchNodes_NoActivity` — no comment → `LatestActivity.Login == ""`

---

### Step 5: Propagate to `issue` and update `toItem()` description

**Files:** `internal/issue/issue.go`, `internal/issue/ui.go`, `internal/issue/issue_test.go`

Add `LatestActivity gh.LatestActivity` to `issue`. Copy it in `fromGraphQL`.

Same fallback logic as Step 3 in `toItem()`.

Tests (`issue_test.go`):
- [x] `TestIssue_ToItem_WithActivity` — description contains `"· commented by alice"`
- [x] `TestIssue_ToItem_NoActivity` — description contains `"updated"` (fallback)
- [x] `TestFromGraphQL_PropagatesLatestActivity` — `fromGraphQL` copies `LatestActivity`

---

## Verification

```sh
go test ./...
go build
gh own pr     # verify PR list shows latest activity
gh own issue  # verify issue list shows latest activity
golangci-lint run
```
