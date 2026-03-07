// Package demo provides hardcoded fake data for recording demos without GitHub API calls.
package demo

import (
	"fmt"

	"github.com/snrsw/gh-own/internal/gh"
)

// PRSearchResult returns a populated fake PRSearchResult for demo use.
func PRSearchResult() *gh.PRSearchResult {
	return &gh.PRSearchResult{
		Created: []gh.PRSearchNode{
			prNode(101, "feat: add dark mode to dashboard",
				"acme-corp/frontend", false, "SUCCESS", "APPROVED",
				gh.LatestActivity{Kind: "approved", Login: "alice", At: "2026-03-06T14:30:00Z"},
				"bob", "2026-03-01T09:00:00Z"),
			prNode(42, "fix: resolve memory leak in worker",
				"acme-corp/backend", false, "FAILURE", "",
				gh.LatestActivity{Kind: "commented", Login: "carol", At: "2026-03-05T11:20:00Z"},
				"bob", "2026-02-25T08:00:00Z"),
			prNode(7, "chore: update CI configuration",
				"demo-org/api-gateway", true, "PENDING", "REVIEW_REQUIRED",
				gh.LatestActivity{},
				"bob", "2026-03-03T16:00:00Z"),
		},
		Assigned: []gh.PRSearchNode{
			prNode(88, "refactor: extract auth middleware",
				"acme-corp/backend", false, "SUCCESS", "CHANGES_REQUESTED",
				gh.LatestActivity{Kind: "changes requested", Login: "dave", At: "2026-03-06T09:00:00Z"},
				"alice", "2026-02-28T10:00:00Z"),
		},
		ReviewRequested: []gh.PRSearchNode{
			prNode(55, "docs: add API usage examples",
				"demo-org/api-gateway", false, "SUCCESS", "REVIEW_REQUIRED",
				gh.LatestActivity{Kind: "pushed", Login: "carol", At: "2026-03-07T08:00:00Z"},
				"carol", "2026-03-04T14:00:00Z"),
		},
		Participated: []gh.PRSearchNode{
			prNode(120, "feat: integrate payment provider",
				"acme-corp/frontend", false, "", "",
				gh.LatestActivity{Kind: "commented", Login: "bob", At: "2026-03-06T17:00:00Z"},
				"alice", "2026-02-20T12:00:00Z"),
		},
		Custom: make(map[string][]gh.PRSearchNode),
	}
}

// IssueSearchResult returns a populated fake IssueSearchResult for demo use.
func IssueSearchResult() *gh.IssueSearchResult {
	return &gh.IssueSearchResult{
		Created: []gh.IssueSearchNode{
			issueNode(301, "bug: login fails on Safari 17",
				"acme-corp/frontend", "OPEN",
				gh.LatestActivity{Kind: "commented", Login: "alice", At: "2026-03-06T10:00:00Z"},
				"bob", "2026-03-01T08:00:00Z"),
			issueNode(202, "feat: support multi-factor authentication",
				"acme-corp/backend", "OPEN",
				gh.LatestActivity{},
				"bob", "2026-02-20T09:00:00Z"),
		},
		Assigned: []gh.IssueSearchNode{
			issueNode(415, "chore: upgrade dependency versions",
				"demo-org/api-gateway", "OPEN",
				gh.LatestActivity{Kind: "commented", Login: "carol", At: "2026-03-05T15:30:00Z"},
				"dave", "2026-02-28T11:00:00Z"),
		},
		Participated: []gh.IssueSearchNode{
			issueNode(188, "docs: improve onboarding guide",
				"acme-corp/frontend", "OPEN",
				gh.LatestActivity{Kind: "commented", Login: "bob", At: "2026-03-07T07:00:00Z"},
				"alice", "2026-02-15T14:00:00Z"),
		},
		Custom: make(map[string][]gh.IssueSearchNode),
	}
}

func prNode(num int, title, repo string, draft bool, ci, review string,
	activity gh.LatestActivity, author, createdAt string) gh.PRSearchNode {
	updatedAt := activity.At
	if updatedAt == "" {
		updatedAt = createdAt
	}
	n := gh.PRSearchNode{
		Number:         num,
		Title:          title,
		URL:            fmt.Sprintf("https://github.com/%s/pull/%d", repo, num),
		IsDraft:        draft,
		StatusState:    ci,
		ReviewDecision: review,
		LatestActivity: activity,
		UpdatedAt:      updatedAt,
		CreatedAt:      createdAt,
	}
	n.Author.Login = author
	n.Repository.NameWithOwner = repo
	return n
}

func issueNode(num int, title, repo, state string,
	activity gh.LatestActivity, author, createdAt string) gh.IssueSearchNode {
	updatedAt := activity.At
	if updatedAt == "" {
		updatedAt = createdAt
	}
	n := gh.IssueSearchNode{
		Number:         num,
		Title:          title,
		URL:            fmt.Sprintf("https://github.com/%s/issues/%d", repo, num),
		State:          state,
		LatestActivity: activity,
		UpdatedAt:      updatedAt,
		CreatedAt:      createdAt,
	}
	n.Author.Login = author
	n.Repository.NameWithOwner = repo
	return n
}
