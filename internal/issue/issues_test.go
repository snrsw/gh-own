package issue

import (
	"strings"
	"testing"

	"github.com/snrsw/gh-own/internal/gh"
)

func TestIssue_RepositoryFullName(t *testing.T) {
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
			repositoryURL: "https://api.github.com/repos/owner/repo/issues/123",
			expected:      "issues/123",
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
			issue := &Issue{
				RepositoryURL: tt.repositoryURL,
			}
			result := issue.RepositoryFullName()
			if result != tt.expected {
				t.Errorf("Issue.RepositoryFullName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestIssue_ToItem(t *testing.T) {
	issue := Issue{
		Number:        42,
		User:          gh.User{Login: "testuser"},
		RepositoryURL: "https://api.github.com/repos/owner/repo",
		Title:         "Fix the bug",
		State:         "open",
		HTMLURL:       "https://github.com/owner/repo/issues/42",
		UpdatedAt:     "2024-03-15T10:30:00Z",
		CreatedAt:     "2024-03-10T08:00:00Z",
	}

	item := issue.toItem()

	// Verify title format: "owner/repo Title"
	expectedTitle := "owner/repo Fix the bug"
	if got := item.Title(); got != expectedTitle {
		t.Errorf("Title() = %q, want %q", got, expectedTitle)
	}

	// Verify description contains expected components
	desc := item.Description()
	expectedParts := []string{"#42", "2024-03-10", "testuser"}
	for _, part := range expectedParts {
		if !strings.Contains(desc, part) {
			t.Errorf("Description() = %q, should contain %q", desc, part)
		}
	}

	// FilterValue should match title for search
	if got := item.FilterValue(); got != expectedTitle {
		t.Errorf("FilterValue() = %q, want %q", got, expectedTitle)
	}
}

func TestGroupedIssues_IssueItems(t *testing.T) {
	tests := []struct {
		name     string
		input    gh.SearchResult[Issue]
		expected int
	}{
		{
			name: "multiple issues",
			input: gh.SearchResult[Issue]{
				TotalCount: 2,
				Items: []Issue{
					{Number: 1, RepositoryURL: "https://api.github.com/repos/owner/repo1", Title: "Issue 1"},
					{Number: 2, RepositoryURL: "https://api.github.com/repos/owner/repo2", Title: "Issue 2"},
				},
			},
			expected: 2,
		},
		{
			name: "empty list",
			input: gh.SearchResult[Issue]{
				TotalCount: 0,
				Items:      []Issue{},
			},
			expected: 0,
		},
		{
			name: "single issue",
			input: gh.SearchResult[Issue]{
				TotalCount: 1,
				Items: []Issue{
					{Number: 42, RepositoryURL: "https://api.github.com/repos/owner/repo", Title: "Solo Issue"},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := &GroupedIssues{Created: tt.input}
			items := grouped.issueItems(grouped.Created)

			if len(items) != tt.expected {
				t.Errorf("issueItems() returned %d items, want %d", len(items), tt.expected)
			}
		})
	}
}
