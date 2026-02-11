package pr

import (
	"strings"
	"testing"

	"github.com/snrsw/gh-own/internal/cistatus"
	"github.com/snrsw/gh-own/internal/gh"
)

func TestPullRequest_RepositoryFullName(t *testing.T) {
	tests := []struct {
		name          string
		repositoryURL string
		expected      string
	}{
		{
			name:          "valid GitHub API URL",
			repositoryURL: "https://api.github.com/repos/owner/repo",
			expected:      "owner/repo",
		},
		{
			name:          "valid URL with longer path",
			repositoryURL: "https://api.github.com/repos/my-org/my-repo",
			expected:      "my-org/my-repo",
		},
		{
			name:          "URL with extra path segments",
			repositoryURL: "https://api.github.com/repos/owner/repo/pulls/123",
			expected:      "pulls/123",
		},
		{
			name:          "empty URL",
			repositoryURL: "",
			expected:      "",
		},
		{
			name:          "short URL with fewer than 5 parts",
			repositoryURL: "https://api.github.com/repos",
			expected:      "",
		},
		{
			name:          "exactly 5 parts",
			repositoryURL: "https://api.github.com/repos/owner/repo",
			expected:      "owner/repo",
		},
		{
			name:          "GitHub Enterprise URL",
			repositoryURL: "https://github.example.com/api/v3/repos/org/project",
			expected:      "org/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := &pullRequest{
				RepositoryURL: tt.repositoryURL,
			}
			result := pr.repositoryFullName()
			if result != tt.expected {
				t.Errorf("PullRequest.repositoryFullName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPullRequest_ToItem(t *testing.T) {
	tests := []struct {
		name          string
		pr            pullRequest
		expectedTitle string
		descContains  []string
	}{
		{
			name: "regular PR",
			pr: pullRequest{
				Number:        123,
				User:          gh.User{Login: "contributor"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Add new feature",
				Draft:         false,
				CreatedAt:     "2024-03-10T08:00:00Z",
			},
			expectedTitle: "owner/repo Add new feature -",
			descContains:  []string{"#123", "2024-03-10", "contributor"},
		},
		{
			name: "draft PR",
			pr: pullRequest{
				Number:        456,
				User:          gh.User{Login: "author"},
				RepositoryURL: "https://api.github.com/repos/org/project",
				Title:         "Work in progress",
				Draft:         true,
				CreatedAt:     "2024-01-15T12:00:00Z",
			},
			expectedTitle: "org/project Work in progress -",
			descContains:  []string{"#456", "2024-01-15", "author"},
		},
		{
			name: "PR with CI status success",
			pr: pullRequest{
				Number:        789,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Feature with CI",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				CIStatus:      cistatus.CIStatusSuccess,
			},
			expectedTitle: "owner/repo Feature with CI ✓",
			descContains:  []string{"#789"},
		},
		{
			name: "PR with CI status failure",
			pr: pullRequest{
				Number:        101,
				User:          gh.User{Login: "dev"},
				RepositoryURL: "https://api.github.com/repos/owner/repo",
				Title:         "Failing CI",
				Draft:         false,
				CreatedAt:     "2024-03-15T10:00:00Z",
				CIStatus:      cistatus.CIStatusFailure,
			},
			expectedTitle: "owner/repo Failing CI ✗",
			descContains:  []string{"#101"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := tt.pr.toItem()

			if got := item.Title(); got != tt.expectedTitle {
				t.Errorf("Title() = %q, want %q", got, tt.expectedTitle)
			}

			desc := item.Description()
			for _, part := range tt.descContains {
				if !strings.Contains(desc, part) {
					t.Errorf("Description() = %q, should contain %q", desc, part)
				}
			}

			if got := item.FilterValue(); got != tt.expectedTitle {
				t.Errorf("FilterValue() = %q, want %q", got, tt.expectedTitle)
			}
		})
	}
}

func TestRenderPRNumber(t *testing.T) {
	tests := []struct {
		name     string
		number   int
		draft    bool
		contains string
	}{
		{
			name:     "regular PR",
			number:   42,
			draft:    false,
			contains: "#42",
		},
		{
			name:     "draft PR",
			number:   99,
			draft:    true,
			contains: "#99",
		},
		{
			name:     "large number",
			number:   12345,
			draft:    false,
			contains: "#12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenderPRNumber(tt.number, tt.draft)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("RenderPRNumber(%d, %v) = %q, should contain %q", tt.number, tt.draft, result, tt.contains)
			}
		})
	}
}

func TestGroupedPullRequests_PRItems(t *testing.T) {
	tests := []struct {
		name     string
		input    gh.SearchResult[pullRequest]
		expected int
	}{
		{
			name: "multiple PRs",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 2,
				Items: []pullRequest{
					{Number: 1, RepositoryURL: "https://api.github.com/repos/owner/repo1", Title: "PR 1"},
					{Number: 2, RepositoryURL: "https://api.github.com/repos/owner/repo2", Title: "PR 2"},
				},
			},
			expected: 2,
		},
		{
			name: "empty list",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 0,
				Items:      []pullRequest{},
			},
			expected: 0,
		},
		{
			name: "single PR",
			input: gh.SearchResult[pullRequest]{
				TotalCount: 1,
				Items: []pullRequest{
					{Number: 42, RepositoryURL: "https://api.github.com/repos/owner/repo", Title: "Solo PR"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := &GroupedPullRequests{Created: tt.input}
			items := grouped.prItems(grouped.Created)

			if len(items) != tt.expected {
				t.Errorf("prItems() returned %d items, want %d", len(items), tt.expected)
			}
		})
	}
}

func TestPullRequest_HasCIStatusField(t *testing.T) {
	pr := pullRequest{
		Number:   123,
		CIStatus: cistatus.CIStatusSuccess,
	}

	if pr.CIStatus != cistatus.CIStatusSuccess {
		t.Errorf("CIStatus = %v, want %v", pr.CIStatus, cistatus.CIStatusSuccess)
	}
}

func TestFromGraphQL(t *testing.T) {
	node := gh.PRSearchNode{
		Number:      123,
		Title:       "Test PR",
		URL:         "https://github.com/owner/repo/pull/123",
		IsDraft:     false,
		UpdatedAt:   "2024-03-10T10:00:00Z",
		CreatedAt:   "2024-03-10T08:00:00Z",
		StatusState: "SUCCESS",
	}
	node.Author.Login = "testuser"
	node.Repository.NameWithOwner = "owner/repo"

	pr := fromGraphQL(node)

	if pr.Number != 123 {
		t.Errorf("Number = %d, want 123", pr.Number)
	}
	if pr.Title != "Test PR" {
		t.Errorf("Title = %q, want %q", pr.Title, "Test PR")
	}
	if pr.User.Login != "testuser" {
		t.Errorf("User.Login = %q, want %q", pr.User.Login, "testuser")
	}
	if pr.CIStatus != cistatus.CIStatusSuccess {
		t.Errorf("CIStatus = %v, want %v", pr.CIStatus, cistatus.CIStatusSuccess)
	}
	if pr.HTMLURL != "https://github.com/owner/repo/pull/123" {
		t.Errorf("HTMLURL = %q, want %q", pr.HTMLURL, "https://github.com/owner/repo/pull/123")
	}
}

func TestFromGraphQLNodes(t *testing.T) {
	nodes := []gh.PRSearchNode{
		{Number: 1, Title: "PR 1", StatusState: "SUCCESS"},
		{Number: 2, Title: "PR 2", StatusState: "FAILURE"},
	}

	prs := fromGraphQLNodes(nodes)

	if len(prs) != 2 {
		t.Fatalf("FromGraphQLNodes returned %d items, want 2", len(prs))
	}
	if prs[0].Number != 1 {
		t.Errorf("prs[0].Number = %d, want 1", prs[0].Number)
	}
	if prs[1].CIStatus != cistatus.CIStatusFailure {
		t.Errorf("prs[1].CIStatus = %v, want %v", prs[1].CIStatus, cistatus.CIStatusFailure)
	}
}
