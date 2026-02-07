# Plan: Add CI Status to PR List

Add a CI status indicator (pass/fail/pending) to each PR displayed in `gh-own pr`.

## Approach

Use **unified GraphQL API** to fetch PRs and CI status together in a single query per search group. This replaces the current REST API search with GraphQL `search` query that includes `statusCheckRollup.state`.

## UI Design

Current description:
```
#123 opened on 2024-03-10 by user, updated 2h ago
```

New description with CI status:
```
#123 ✓ opened on 2024-03-10 by user, updated 2h ago
```

| Status  | Icon | Color   |
|---------|------|---------|
| Success | ✓    | #1A7F37 |
| Failure | ✗    | #CF222E |
| Pending | ●    | #9A6700 |
| None    | -    | (gray)  |

## GraphQL Query

```graphql
query($q: String!) {
  search(query: $q, type: ISSUE, first: 50) {
    nodes {
      ... on PullRequest {
        number
        title
        author { login }
        repository { nameWithOwner }
        url
        isDraft
        updatedAt
        createdAt
        commits(last: 1) {
          nodes {
            commit {
              statusCheckRollup { state }
            }
          }
        }
      }
    }
  }
}
```

## Implementation (TDD)

### Test 1: CIStatus type and String method
- [x] Create `internal/cistatus/cistatus.go` with `CIStatus` type
- [x] Constants: `CIStatusSuccess`, `CIStatusFailure`, `CIStatusPending`, `CIStatusNone`
- [x] `String()` method for each status

### Test 2: ParseState converts GraphQL state to CIStatus
- [x] `ParseState("SUCCESS")` → `CIStatusSuccess`
- [x] `ParseState("FAILURE")` / `"ERROR"` → `CIStatusFailure`
- [x] `ParseState("PENDING")` / `"EXPECTED"` → `CIStatusPending`
- [x] `ParseState("")` → `CIStatusNone`

### Test 3: RenderCIStatus returns styled icon
- [x] Add `RenderCIStatus(status CIStatus) string` in `internal/cistatus/cistatus.go`
- [x] Return styled icons with appropriate colors

### Test 4: PullRequest struct has CIStatus field
- [x] Add `CIStatus cistatus.CIStatus` field to `PullRequest` struct in `internal/pullrequest/pullrequests.go`

### Test 5: toItem includes CI status in description
- [x] Update `toItem()` in `internal/pullrequest/ui.go` to include CI status icon after PR number

### Test 6: GraphQL search returns PRs with CI status
- [x] Add `internal/gh/graphql.go` with GraphQL client helper
- [x] Add `internal/gh/search_graphql.go` with `SearchGraphQL` function
- [x] Parallel queries using goroutines (same pattern as REST `Search`)

### Test 7: Update SearchPullRequests to use GraphQL
- [x] Add `SearchPullRequestsGraphQL` function using `gh.SearchGraphQL`
- [x] Add `FromGraphQL` and `FromGraphQLNodes` to map GraphQL response to `PullRequest`
- [x] Add `BuildConditions` helper to share conditions between REST and GraphQL

### Test 8: Wire up GraphQL client in cmd/pr.go
- [x] Update `CurrentUser()` to return GraphQL client
- [x] Add `CurrentUserREST()` for backward compatibility with issues
- [x] Update `SearchPullRequests` to use GraphQL client

## Files to Modify/Create

| File | Action |
|------|--------|
| `internal/cistatus/cistatus.go` | Create - CIStatus type, constants, parsing, rendering |
| `internal/cistatus/cistatus_test.go` | Create - Unit tests |
| `internal/gh/graphql.go` | Create - GraphQL client helper |
| `internal/gh/search_graphql.go` | Create - GraphQL search function |
| `internal/gh/search_graphql_test.go` | Create - GraphQL search tests |
| `internal/pullrequest/pullrequests.go` | Modify - Add CIStatus field, use GraphQL |
| `internal/pullrequest/ui.go` | Modify - Include CI status in toItem |
| `cmd/pr.go` | Modify - Use GraphQL client |

## Verification

1. Run `go test ./...` after each test implementation
2. Run `go build` to verify compilation
3. Run `gh own pr` manually to verify CI status icons appear
4. Test with PRs that have: passing CI, failing CI, pending CI, no CI
